#!/bin/bash

set -e

usage() {
  echo
  echo "Usage: $0 [-h] [-v]"
  echo "Options:"
  echo "-h Show this screen"
  echo "-v rpm package version"
  echo
}

build() {
    local package=$1
    local version=$2
      local package_name="orcus-exporter"
    rm -rf dist/rpm_build
    mkdir -p dist/rpm_build
    mkdir -p dist/rpm_build/var/log/orcus-exporter
    mkdir -p dist/rpm_build/usr/bin
    mkdir -p dist/rpm_build/etc/systemd/system
    cp etc/systemd/orcus-exporter.service dist/rpm_build/etc/systemd/system
    cp dist/orcus-exporter_linux_amd64/orcus-exporter dist/rpm_build/usr/bin

    fpm -v $version --epoch 1 -f -s dir -n $package_name -m "MaxFedotov <m.a.fedotov@gmail.com>" --description "Orcus Prometheus Exporter" --url "https://github.com/MaxFedotov/orcus-exporter" --vendor "MaxFedotov" --license "Apache 2.0" --rpm-os linux --before-install scripts/pre-install.sh --rpm-attr 744,orcus_exporter,orcus_exporter:/var/log/orcus-exporter -t rpm -C dist/rpm_build -p dist/

    rm -rf dist/rpm_build
}

while getopts "s:p:v:h" opt; do
  case $opt in
  s)
    source="${OPTARG}"
    ;;
  h)
    usage
    exit 0
    ;;
  p)
    package="${OPTARG}"
    ;;
  v)
    version="${OPTARG}"
    ;;
  ?)
    usage
    exit 2
    ;;
  esac
done

shift $(( OPTIND - 1 ));

if [ -z "$version" ]; then
    echo "Error. Version not specified"
    exit 1
fi

build "systemd" "$version"

