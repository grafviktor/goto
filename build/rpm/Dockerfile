
# Verify the package has been installed
# rpm -qi goto
#
# See changelog
# rpm -q goto --changelog
#
# See what’s in the RPM package
# rpm -ql goto
#
# Remove the RPM package
# dnf remove goto
# rpm --verbose --erase hello

FROM almalinux:9.3-20231124 as build-stage

ARG VERSION
RUN yum install -y rpmdevtools rpmlint golang git
ADD build/rpm/goto.spec /tmp/
RUN rpmdev-setuptree && \
  # rpmbuild should use XZ compression for packages, to avoid "rpmlib(PayloadIsZstd) <= 5.4.18-1" is needed error
  echo '%_binary_payload w2.xzdio' > ~/.rpmmacros && \
  cp /tmp/goto.spec ~/rpmbuild/SPECS/ && \
  # spectool is used to download source code
  spectool -gR --define "_version $VERSION" ~/rpmbuild/SPECS/goto.spec && \
  rpmbuild -ba --define "_version $VERSION" --define "_date $(date)" ~/rpmbuild/SPECS/goto.spec && \
  rpm -ivh ~/rpmbuild/RPMS/x86_64/goto-${VERSION}.x86_64.rpm && \
  gg -v

# Copy from build container to host file system.
# Ensure that DOCKER_BUILDKIT=1 and docker build has parameter: --output build/dist
FROM scratch AS export-stage
ARG VERSION
COPY --from=build-stage /root/rpmbuild/RPMS/x86_64/goto-${VERSION}.x86_64.rpm .