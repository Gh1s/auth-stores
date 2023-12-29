#!/bin/bash

while getopts 'spr:' flag;
do
  case "${flag}" in
    s) stable="true"
       echo "Building with stable tag" ;;
    p) push="true"
       echo "Pushing after build";;
    r) IFS="," read -ra additional_registries <<< "${OPTARG}"
       echo "Additionnal registries: ${additional_registries[@]}";;
    *) echo "Unknown parameter passed: $1"; exit 1 ;;
  esac
done

LATEST_TAG=latest
STABLE_TAG=stable

export BUILD_CONTEXT=../
export REGISTRY=${REGISTRY:-gcr.io/csb-anthos}
export GRPC_LDAP_REPOSITORY=auth/stores/ldap
export GRPC_ACCOUNTS_REPOSITORY=auth/stores/accounts
export GRPC_ACCOUNTS_TAG=$LATEST_TAG
export GRPC_LDAP_TAG=$LATEST_TAG
echo "Building images with the default '$LATEST_TAG' tag."
docker-compose -f docker-compose-build.yaml build

if [ "$stable" == "true" ]
  then
    echo "Tagging images with the the 'stable' tag."
    docker tag $REGISTRY/$GRPC_LDAP_REPOSITORY:$LATEST_TAG $REGISTRY/$GRPC_LDAP_REPOSITORY:$STABLE_TAG
    docker tag $REGISTRY/$GRPC_ACCOUNTS_REPOSITORY:$LATEST_TAG $REGISTRY/$GRPC_ACCOUNTS_REPOSITORY:$STABLE_TAG
fi

export GRPC_LDAP_VERSION=$(cat $BUILD_CONTEXT/grpc/ldap/go.mod.version)
export GRPC_ACCOUNTS_VERSION=$(cat $BUILD_CONTEXT/grpc/ldap/go.mod.version)
echo "Tagging images with their version tag."
docker tag $REGISTRY/$GRPC_LDAP_REPOSITORY:$LATEST_TAG $REGISTRY/$GRPC_LDAP_REPOSITORY:$GRPC_LDAP_VERSION
docker tag $REGISTRY/$GRPC_ACCOUNTS_REPOSITORY:$LATEST_TAG $REGISTRY/$GRPC_ACCOUNTS_REPOSITORY:$GRPC_ACCOUNTS_VERSION

if [ "${#additional_registries[@]}" -gt "0" ]
  echo "Tagging images with the additional registries."
  then
    for additional_registry in "${additional_registries[@]}"; do
      docker tag $REGISTRY/$GRPC_LDAP_REPOSITORY:$LATEST_TAG $additional_registry/$GRPC_LDAP_REPOSITORY:$LATEST_TAG
      docker tag $REGISTRY/$GRPC_ACCOUNTS_REPOSITORY:$LATEST_TAG $additional_registry/$GRPC_ACCOUNTS_REPOSITORY:$LATEST_TAG
      if [ "$stable" == "true" ]
        then
          docker tag $REGISTRY/$GRPC_LDAP_REPOSITORY:$LATEST_TAG $additional_registry/$GRPC_LDAP_REPOSITORY:$STABLE_TAG
          docker tag $REGISTRY/$GRPC_ACCOUNTS_REPOSITORY:$LATEST_TAG $additional_registry/$GRPC_ACCOUNTS_REPOSITORY:$STABLE_TAG
      fi
      docker tag $REGISTRY/$GRPC_LDAP_REPOSITORY:$LATEST_TAG $additional_registry/$GRPC_LDAP_REPOSITORY:$GRPC_LDAP_VERSION
      docker tag $REGISTRY/$GRPC_ACCOUNTS_REPOSITORY:$LATEST_TAG $additional_registry/$GRPC_ACCOUNTS_REPOSITORY:$GRPC_ACCOUNTS_VERSION
    done
fi

if [ "$push" == "true" ]
then
  echo "Pushing images."
  docker push $REGISTRY/$GRPC_LDAP_REPOSITORY --all-tags
  docker push $REGISTRY/$GRPC_ACCOUNTS_REPOSITORY --all-tags
fi