FROM debian:12.4 as deb-build-stage

ARG VERSION
ENV CGO_ENABLED=0
ENV WORK_DIR="/build"
ENV PACKAGE_SRC_ROOT="${WORK_DIR}/goto_${VERSION}_x86_64"

RUN apt update && apt install git build-essential golang-go -y
ADD build/deb/goto.control build/deb/preinst /tmp/
WORKDIR "$WORK_DIR"
RUN \
  git clone --depth 1 --branch v%{_version} https://github.com/grafviktor/goto.git . && \
  make build && \
  mkdir -p ${PACKAGE_SRC_ROOT}/DEBIAN && \
  cp /tmp/preinst ${PACKAGE_SRC_ROOT}/DEBIAN && \
  cp /tmp/goto.control ${PACKAGE_SRC_ROOT}/DEBIAN/control && \
  mkdir -p ${PACKAGE_SRC_ROOT}/usr/bin && \
  cp ./build/dist/gg ${PACKAGE_SRC_ROOT}/usr/bin && \
  cd "$PACKAGE_SRC_ROOT" && \
  dpkg-deb --build && \
  dpkg -i goto-${VERSION}.x86_64.deb && \
  gg -v

FROM scratch AS export-stage
ARG VERSION
COPY --from=deb-build-stage /root/rpmbuild/RPMS/x86_64/goto-${VERSION}.x86_64.rpm .