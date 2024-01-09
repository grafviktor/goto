Name:           goto
Version:        0.4.0
Release:        1%{?dist}
Summary:        GOTO - A simple command line SSH manager
ExclusiveArch:  x86_64
Group:          Applications/System

License:        MIT
Source0:        https://github.com/grafviktor/goto/archive/refs/tags/v0.4.0.tar.gz
# Source0:        goto-0.4.0.tar.gz

%define _build_id_links none
%global debug_package %{nil}

%description
This utility helps to maintain a list of ssh servers. Unlike PuTTY it doesn't incorporate any connection logic, but relying on ssh utility which should be installed on your system.

%prep
%setup -q #unpack tarball
%build

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/%{_bindir}
cp gg $RPM_BUILD_ROOT/%{_bindir}


%clean
rm -rf $RPM_BUILD_ROOT

%files
%{_bindir}/gg

%changelog
* Thu Jan  04 2024 Roman Leonenkov <some@mail> - 0.4.0
- First version being packaged