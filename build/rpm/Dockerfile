
# https://rpm-packaging-guide.github.io/
# https://www.redhat.com/sysadmin/create-rpm-package

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


FROM almalinux:9.3-20231124

ARG VERSION=0.4.0
WORKDIR /tmp

RUN dnf install -y rpmdevtools rpmlint && \
  rpmdev-setuptree && \
  mkdir goto-$VERSION && \
  cp build/dist/gg-lin goto-$VERSION/gg && \
  # tar --create --file goto-$VERSION.tar.gz goto-$VERSION && \
  tar --create --file goto-$VERSION.tar.gz goto-$VERSION && \
  cp goto-$VERSION.tar.gz ~/rpmbuild/SOURCES && \
  cp build/rpm/goto.spec ~/rpmbuild/SPECS/ && \
  rpmlint ~/rpmbuild/SPECS/goto.spec && \
  # download source code
  # spectool -gR ~/rpmbuild/SPECS/goto.spec && \
  rpmbuild -bb ~/rpmbuild/SPECS/goto.spec && \
  ls ~/rpmbuild