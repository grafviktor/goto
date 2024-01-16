
# Install
# dnf install ~/rpmbuild/RPMS/x86_64/goto-0.4.0-1.el9.x86_64.rpm
# rpm -ivh ~/rpmbuild/RPMS/x86_64/goto-0.4.0-1.el9.x86_64.rpm
#
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
ADD build/rpm/goto.spec /tmp/
RUN dnf install -y rpmdevtools rpmlint golang git && \
  rpmdev-setuptree && \
  cp /tmp/goto.spec ~/rpmbuild/SPECS/ && \
  # spectool is used to download source code
  spectool -gR --define "_version $VERSION" ~/rpmbuild/SPECS/goto.spec && \
  rpmbuild -ba --define "_version $VERSION" --define "_date $(date)" ~/rpmbuild/SPECS/goto.spec && \
  rpm -ivh ~/rpmbuild/RPMS/x86_64/goto-${VERSION}-1.el9.x86_64.rpm && \
  gg -v

# Copy from build container to host file system.
# Ensure that DOCKER_BUILDKIT=1 and docker build has parameter: --output build/dist
FROM scratch AS export-stage
ARG VERSION
COPY --from=build-stage /root/rpmbuild/RPMS/x86_64/goto-${VERSION}-1.el9.x86_64.rpm .