#!/bin/sh

VERSION=$(git describe)

mkdir -p dist
cp work-env dist
strip dist/work-env

cd dist && \
    for arch in deb rpm; do
        fpm \
        --input-type dir \
        --output-type $arch \
        --force \
        --name work-env \
        --license MIT \
        --version ${VERSION} \
        --architecture native \
        --vendor 'Andrei Kopats' \
        --maintainer 'andrei.kopats@gmail.com' \
        --description 'Command line environment in Docker for developers' \
        --url 'https://github.com/andreikop/work-env' \
        --prefix /usr/bin/ \
        work-env; \
    done
