# Build stage
FROM debian:trixie-20250224 as deb-build-stage
ARG VERSION
ARG BRANCH
ENV LANG=C
ENV CGO_ENABLED=0
ENV WORK_DIR="/build"
ENV PACKAGE_SRC_ROOT="${WORK_DIR}/goto_${VERSION}_amd64"
RUN apt update >/dev/null && apt install git build-essential golang-go -y >/dev/null
WORKDIR "$WORK_DIR"
RUN \
  git clone --depth 1 --branch ${BRANCH} https://github.com/ts9744/2025-OSP.git . && \
  make build
ADD build/deb/goto.control /tmp/
RUN \
  mkdir -p ${PACKAGE_SRC_ROOT}/DEBIAN && \
  cp /tmp/goto.control ${PACKAGE_SRC_ROOT}/DEBIAN/control && \
  sed -i "s/%VERSION%/$VERSION/g" ${PACKAGE_SRC_ROOT}/DEBIAN/control && \
  sed -i "s/%DATE%/$(date '+%a %b %d %Y')/g" ${PACKAGE_SRC_ROOT}/DEBIAN/control && \
  mkdir -p ${PACKAGE_SRC_ROOT}/usr/bin && \
  cp ./dist/gg ${PACKAGE_SRC_ROOT}/usr/bin && \
  dpkg-deb -v --build "${WORK_DIR}/goto_${VERSION}_amd64"
RUN \
  dpkg -i goto_${VERSION}_amd64.deb && \
  echo -e "\033[0;92mDEB build complete!\033[0m" && \
  echo -e "\033[0;92m$(gg -v)\033[0m"


# Export stage
FROM scratch AS export-stage
ARG VERSION
ENV WORK_DIR="/build"
COPY --from=deb-build-stage ${WORK_DIR}/goto_${VERSION}_amd64.deb .
