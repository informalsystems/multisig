FROM registry.access.redhat.com/ubi8/go-toolset:1.17.12-3 AS go-builder

#ENV GOBIN=/root/go/bin

# Building gaia
RUN git clone https://github.com/cosmos/gaia
WORKDIR gaia
RUN git checkout v7.0.2
RUN make install

# Minio client is needed to configure user/bucket
RUN go install github.com/minio/mc@latest

# Building multisig
RUN mkdir multisig
COPY . multisig
WORKDIR multisig
RUN go install

FROM minio/minio:RELEASE.2022-08-11T04-37-28Z

# Prerequisite
RUN microdnf install git jq
ENV HOME=/root
ENV PATH=$PATH:$HOME/go/bin

# Pulling test framework
WORKDIR $HOME
RUN git clone https://github.com/bats-core/bats-core
RUN git clone https://github.com/bats-core/bats-support
RUN git clone https://github.com/bats-core/bats-assert

# Copy multisig files
WORKDIR "$HOME"
RUN mkdir multisig
COPY . multisig

# Copy of built gaia, mc, multiisg
COPY --from=go-builder /opt/app-root/src/go/bin /root/go/bin

# Configure gaia and add accounts to keychain
WORKDIR "$HOME"
RUN /bin/bash ./multisig/tests/configure_gaia.sh

ENTRYPOINT ["/bin/bash"]
CMD ["./multisig/tests/start.sh"]
