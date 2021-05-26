#
# This is work-env environment image which can be used to build work-env .deb and .rpm packages.
# Usage:
#       work-env build . workenv-build-env
#       work-env run build-env workenv-build-env
#       cd ..
#       ./build-packages
#

FROM work-env

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install ruby ruby-dev rubygems build-essential rpm

RUN gem install fpm
