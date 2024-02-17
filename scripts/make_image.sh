. ./.env
docker build --tag ${IMAGE_NAME}:latest --tag ${IMAGE_NAME}:${IMAGE_VERSION} -f build/package/Dockerfile .
docker image save --output=${IMAGE_FOLDER}/${IMAGE_NAME}:${IMAGE_VERSION} ${IMAGE_NAME}:${IMAGE_VERSION}