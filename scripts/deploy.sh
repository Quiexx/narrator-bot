. ./.env

echo "Save image '${IMAGE_NAME}:${IMAGE_VERSION}' as '$RESULT_IMAGE_FILENAME.tar.gz' archive..."
docker save ${IMAGE_NAME}:${IMAGE_VERSION} | gzip > $RESULT_IMAGE_FILENAME.tar.gz
echo "Image archive created with name $RESULT_IMAGE_FILENAME.tar.gz"

rsync -P ${RESULT_IMAGE_FILENAME}.tar.gz ${SERVER_USER}@${SERVER_IP}:${SERVER_PATH}