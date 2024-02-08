# Warning: you run "echo '%_binary_payload w2.xzdio' > ~/.rpmmacros" before running rpmbuild. #
# That should be done to use an old compression method to avoid newer rpmlib dependency.      #

Name:          goto
Version:       %{_version}
Release:       1%{?dist}
Summary:       GOTO - A simple command line SSH manager
ExclusiveArch: x86_64
BuildRequires: golang
Group:         Applications/System
License:       MIT
Source0:       https://github.com/grafviktor/goto/archive/refs/tags/v%{_version}.tar.gz

# Define a file name without using RH release version in resulting package suffix
%define _rpmfilename %%{ARCH}/%{NAME}-%%{VERSION}.%%{ARCH}.rpm
%define _build_id_links none
%global debug_package %{nil}

%description
This utility helps to maintain a list of ssh servers. Unlike PuTTY it doesn't incorporate any connection logic, but relying on openssh package which should be installed on your system.

%prep
rm -rf rpmbuild/BUILD/{,.[!.],..?}*
git clone --depth 1 --branch v%{_version} https://github.com/grafviktor/goto.git .

%build
# To avoid clib dependency and make this package portable across distributions, disable cgo
export CGO_ENABLED=0
make build

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/%{_bindir}
cp ./dist/gg $RPM_BUILD_ROOT/%{_bindir}

%clean
rm -rf $RPM_BUILD_ROOT

%files
%{_bindir}/gg

%changelog
* %{_date} Roman Leonenkov <6890447+grafviktor@users.noreply.github.com> - %{_version}
Find full changelog in the project's readme file - https://github.com/grafviktor/goto/blob/master/README.md#5-changelog
