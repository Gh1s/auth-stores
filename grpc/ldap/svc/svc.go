package svc

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"csb.nc/auth/stores/grpc/ldap/guid"
	"csb.nc/auth/stores/users"
	"github.com/go-ldap/ldap"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	// LdapConnectionFailed indicates that an error has occured while opening the connection with the LDAP controller.
	LdapConnectionFailed = iota + 1
	// LdapSearchFailed indicates that an error has occured while executing an LDAP query.
	LdapSearchFailed
	// UserNotFound indicates that the requested user has not been found in the LDAP directory.
	UserNotFound
	// UserBindFailed indicates that the LDAP controller binding failed for the provided credentials.
	UserBindFailed
	// UserAccountDisabled indicates that the user's account is disabled.
	UserAccountDisabled
	// UserAccountLocked indicates that the user's account is disabled.
	UserAccountLocked

	viperKeyLdapUsername              = "ldap.username"
	viperKeyLdapPassword              = "ldap.password"
	viperKeyLdapProtocol              = "ldap.protocol"
	viperKeyLdapTLSEnabled            = "ldap.tls.enabled"
	viperKeyLdapTLSRootCAs            = "ldap.tls.rootCAs"
	viperKeyLdapTLSInsecureSkipVerify = "ldap.tls.insecureSkipVerify"
	viperKeyLdapServer                = "ldap.server"
	viperKeyLdapPort                  = "ldap.port"
	viperKeyLdapContainer             = "ldap.container"
	viperKeyLdapAttributes            = "ldap.attributes"
	viperKeyLdapClaimsMapping         = "ldap.claims.mapping"
	viperKeyLdapClaimsChildren        = "ldap.claims.children"

	ldapObjectGUIDFilter       = "(&(objectCategory=person)(objectClass=user)(objectGUID=%s))"
	ldapSAMAccountNameFilter   = "(&(objectCategory=person)(objectClass=user)(sAMAccountName=%s))"
	ldapSearchFilter           = "(&(objectCategory=person)(objectClass=user)(|(sAMAccountName=%[1]s*)(sn=%[1]s*)(givenName=%[1]s*)(displayName=%[1]s*)(mail=%[1]s*)))"
	ldapObjectGUIDAttr         = "objectGUID"
	ldapDnAttr                 = "dn"
	ldapUserAccountControlAttr = "userAccountControl"

	ldapUserAccountControlFlagAccountDisable = 0x00000002
	ldapUserAccountControlFlagLockout        = 0x00000010

	missingLdapCredentialsError = "Could not read LDAP credentials. Please define 'LDAP_USERNAME' and 'LDAP_PASSWORD' environment variables."
)

// Opens the LDAP connection and binds with the domain controller.
func openConn() (*ldap.Conn, error) {
	if viper.GetString(viperKeyLdapUsername) == "" || viper.GetString(viperKeyLdapPassword) == "" {
		zap.L().Error(missingLdapCredentialsError)
	}

	protocol := viper.GetString(viperKeyLdapProtocol)
	tlsEnabled := viper.GetBool(viperKeyLdapTLSEnabled)
	tlsRootCAs := viper.GetStringSlice(viperKeyLdapTLSRootCAs)
	tlsInsecureSkipVerify := viper.GetBool(viperKeyLdapTLSInsecureSkipVerify)
	server := viper.GetString(viperKeyLdapServer)
	port := viper.GetInt(viperKeyLdapPort)

	var conn *ldap.Conn
	var err error

	if tlsEnabled {
		certPool := x509.NewCertPool()
		for _, certFile := range tlsRootCAs {
			if pem, err := ioutil.ReadFile(certFile); err != nil {
				return nil, err
			} else {
				if !certPool.AppendCertsFromPEM(pem) {
					return nil, errors.New("Failed to append CA certificate " + certFile)
				}
			}
		}
		conn, err = ldap.DialTLS(
			protocol,
			fmt.Sprintf("%s:%d", server, port),
			&tls.Config{
				RootCAs:            certPool,
				InsecureSkipVerify: tlsInsecureSkipVerify,
			},
		)
	} else {
		conn, err = ldap.Dial(
			protocol,
			fmt.Sprintf("%s:%d", server, port),
		)
	}

	if err != nil {
		return nil, err
	}

	userName := viper.GetString(viperKeyLdapUsername)
	password := viper.GetString(viperKeyLdapPassword)
	if err := conn.Bind(userName, password); err != nil {
		return nil, err
	}

	return conn, nil
}

// Searches & finds the LDAP attributes values into the LDAP directory.
func findItems(conn *ldap.Conn, filter string, attrs []string) ([]map[string]string, error) {
	req := ldap.NewSearchRequest(
		viper.GetString(viperKeyLdapContainer),
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		attrs,
		nil,
	)
	res, err := conn.Search(req)
	if err != nil {
		// If the search has failed, there's no point to continue.
		return make([]map[string]string, 0), err
	}
	items := make([]map[string]string, len(res.Entries))
	convert := viper.Sub(viperKeyLdapAttributes)

	for index, entry := range res.Entries {
		itemValues := make(map[string]string, len(attrs))

		for _, attr := range attrs {
			// Using the conversion mapping, we try to convert the LDAP attribute value to a human readable string.
			cType := convert.GetString(attr)
			switch cType {
			case "string":
				itemValues[attr] = entry.GetAttributeValue(attr)
			case "guid":
				var bytes [16]byte
				copy(bytes[:], entry.GetRawAttributeValue(attr)[:16])
				itemValues[attr] = guid.FromWindowsArray(bytes).String()
			default:
				// There is nothing we can do for that case.
				zap.L().Sugar().Warnf("Unsupported conversion type: %s", cType)
			}
		}

		itemValues[ldapDnAttr] = entry.DN
		items[index] = itemValues
	}

	return items, nil
}

// Maps claims names to LDAP attributes names
func mapClaimsToLdapAttrs(claims []string) []string {
	mapping := viper.Sub(viperKeyLdapClaimsMapping)
	attrs := make([]string, 0)
	for _, claim := range claims {
		attr := mapping.GetString(claim)
		if attr != "" {
			attrs = append(attrs, attr)
		}
	}
	return attrs
}

// Reserve lookup in the map to find the claim name matching the LDAP attribute name.
func findClaimName(mapping *viper.Viper, attr string) string {
	for key, val := range mapping.AllSettings() {
		if val == attr {
			return key
		}
	}
	return ""
}

// Maps LDAP attributes with value to claims with value.
func mapLdapAttrsToClaims(attrs map[string]string) map[string]string {
	mapping := viper.Sub(viperKeyLdapClaimsMapping)
	children := viper.Sub(viperKeyLdapClaimsChildren)
	claims := make(map[string]string, len(attrs))

	for key, val := range attrs {
		c := findClaimName(mapping, key)
		if c != "" {
			claims[c] = val
			childName := children.GetString(fmt.Sprintf("%s.name", c))
			childValue := children.GetString(fmt.Sprintf("%s.value", c))
			if childName != "" {
				claims[childName] = childValue
			}
		}
	}
	return claims
}

// Authenticate authenticates a user against the domain controler using the provided credentials.
func Authenticate(req *users.AuthRequest) *users.AuthResponse {
	zap.L().Sugar().Infof("Authenticating user: %s", req.Username)

	resp := &users.AuthResponse{}

	zap.L().Debug("Opening LDAP connection.")
	conn, err := openConn()
	if err != nil {
		zap.L().Error("Could not open LDAP connection", zap.Error(err))
		resp.Error = LdapConnectionFailed
		return resp
	}
	defer conn.Close()

	zap.L().Sugar().Debugf("Searching the distinguished name of the user: %s", req.Username)

	filter := fmt.Sprintf(ldapSAMAccountNameFilter, req.Username)
	items, err := findItems(conn, filter, []string{ldapObjectGUIDAttr, ldapUserAccountControlAttr})

	if err != nil {
		zap.L().Error(
			fmt.Sprintf("Could not search LDAP attributes using sAMAccountName: %s", req.Username),
			zap.Error(err),
			zap.String("filter", filter),
			zap.String("userName", req.Username),
		)
		resp.Error = LdapSearchFailed
		return resp
	}

	if len(items) == 0 {
		resp.Error = UserNotFound
		return resp
	}

	zap.L().Debug("Checking if the account is disabled or locked.")
	item := items[0]
	userAccountControlValue := item[ldapUserAccountControlAttr]
	if userAccountControl, err := strconv.ParseInt(userAccountControlValue, 0, 32); err == nil {
		if userAccountControl|ldapUserAccountControlFlagAccountDisable == userAccountControl {
			resp.Error = UserAccountDisabled
		} else if userAccountControl|ldapUserAccountControlFlagLockout == userAccountControl {
			resp.Error = UserAccountLocked
		}
	} else {
		zap.L().Error(
			"Could not parse the user account control flag.",
			zap.Error(err),
			zap.String("userAccountControlValue", userAccountControlValue),
		)
		resp.Error = LdapSearchFailed
	}
	if resp.Error > 0 {
		return resp
	}

	dn := item[ldapDnAttr]
	zap.L().Sugar().Debugf("Binding to the domain controller using distinguished name: %s", dn)

	err = conn.Bind(dn, req.Password)
	if err != nil {
		zap.L().Warn(
			"Domain controller binding failed",
			zap.Error(err),
			zap.String("dn", dn),
			zap.String("userName", req.Username),
		)
		resp.Error = UserBindFailed
		return resp
	}

	resp.Succeeded = true
	resp.Subject = item[ldapObjectGUIDAttr]

	return resp
}

// FindClaims finds the requested claims with the provided identifier.
func FindClaims(req *users.ClaimsRequest) *users.ClaimsResponse {
	zap.L().Sugar().Infof("Searching claims for the user: %d:%s", req.IdentifierType, req.Identifier)

	resp := &users.ClaimsResponse{
		Succeeded: false,
		Claims:    make(map[string]string, len(req.Claims)),
	}

	zap.L().Debug("Opening LDAP connection.")
	conn, err := openConn()
	if err != nil {
		zap.L().Error("Could not open LDAP connection", zap.Error(err))
		resp.Error = LdapConnectionFailed
		return resp
	}
	defer conn.Close()

	var filter string
	switch req.IdentifierType {
	case users.IdentifierType_SUBJECT:
		id, err := guid.FromString(req.Identifier)
		if err != nil {
			zap.L().Error(
				"Could not parse the identifier into a GUID.",
				zap.Error(err),
				zap.String("identifier", req.Identifier),
			)
		}
		// Active Directory requires the objectGUID to be hex string with each hex escaped.
		var sb strings.Builder
		src := make([]byte, 1)
		var dst []byte
		for _, b := range id.ToWindowsArray() {
			src[0] = b
			dst = make([]byte, hex.EncodedLen(len(src)))
			hex.Encode(dst, src)
			sb.WriteString("\\")
			sb.Write(dst)
		}
		filter = fmt.Sprintf(ldapObjectGUIDFilter, sb.String())
	case users.IdentifierType_USER_NAME:
		filter = fmt.Sprintf(ldapSAMAccountNameFilter, req.Identifier)
	default:
		zap.L().Sugar().Warnf("Unsupported identifier type: %d", req.IdentifierType)
		resp.Error = LdapSearchFailed
		return resp
	}

	attrs := mapClaimsToLdapAttrs(req.Claims)
	items, err := findItems(conn, filter, attrs)
	if err != nil {
		zap.L().Error(
			"An error has occured while fetching claims.",
			zap.Error(err),
			zap.String("identifier", req.Identifier),
			zap.Int("identifierType", int(req.IdentifierType)),
		)
		resp.Error = LdapSearchFailed
		return resp
	}

	if len(items) == 0 {
		resp.Error = UserNotFound
		return resp
	}

	item := items[0]
	resp.Succeeded = true
	resp.Claims = mapLdapAttrsToClaims(item)

	return resp
}

// SearchClaims searches the claims with the provided search filter.
func SearchClaims(req *users.SearchRequest) *users.SearchResponse {
	zap.L().Sugar().Infof("Searching though the directory using: %s", req.Search)

	resp := &users.SearchResponse{
		Succeeded: false,
	}

	zap.L().Debug("Opening LDAP connection.")
	conn, err := openConn()
	if err != nil {
		zap.L().Error("Could not open LDAP connection", zap.Error(err))
		resp.Error = LdapConnectionFailed
		return resp
	}
	defer conn.Close()

	filter := fmt.Sprintf(ldapSearchFilter, req.Search)
	zap.L().Sugar().Debugf("LDAP filter: %s", filter)

	attrs := mapClaimsToLdapAttrs(req.Claims)
	items, err := findItems(conn, filter, attrs)
	if err != nil {
		zap.L().Error(
			"An error has occured while fetching claims.",
			zap.Error(err),
			zap.String("search", req.Search),
			zap.Strings("claims", req.Claims),
		)
		resp.Error = LdapSearchFailed
		return resp
	}

	resp.Results = make([]*users.SearchResponseResult, len(items))
	for index, item := range items {
		result := &users.SearchResponseResult{
			Properties: mapLdapAttrsToClaims(item),
		}
		resp.Results[index] = result
	}
	resp.Succeeded = true

	return resp
}
