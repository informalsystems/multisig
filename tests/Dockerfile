# Took minio container of minio because it's already good;
# might change later to base ubi8/alpine and build minio manually
FROM minio/minio

# Prerequencies for installation
RUN microdnf install go git make which jq
ENV HOME=/root
ENV PATH=$PATH:$HOME/go/bin

# Minio client is needed to configure user/bucket
RUN go install github.com/minio/mc@latest

# Clone and install last gaia version
WORKDIR $HOME
RUN ["git", "clone", "https://github.com/cosmos/gaia"]
WORKDIR gaia
RUN ["make", "install"]
# Cleanup gaia sources
WORKDIR $HOME
RUN ["/bin/bash", "rm", "-rf", "gaia"]

WORKDIR $HOME
RUN ["git", "clone", "https://github.com/bats-core/bats-core"]
RUN ["git", "clone", "https://github.com/bats-core/bats-support"]
RUN ["git", "clone", "https://github.com/bats-core/bats-assert"]

# Copy files
WORKDIR $HOME
RUN ["mkdir", "multisig"]
COPY [".", "multisig"]

# Configure gaia and add accounts to keychain
WORKDIR $HOME
RUN ["/bin/bash", "./multisig/tests/configure_gaia.sh"]

WORKDIR multisig
RUN ["go", "install"]

WORKDIR $HOME
ENTRYPOINT ["/bin/bash"]
CMD ["./multisig/tests/start.sh"]
