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

# Configure gaia and add accounts to keychain
WORKDIR $HOME
COPY ["tests/configure_gaia.sh", "."]
RUN ["/bin/bash", "./configure_gaia.sh"]

WORKDIR $HOME
RUN ["git", "clone", "https://github.com/bats-core/bats-core"]
RUN ["git", "clone", "https://github.com/bats-core/bats-support"]
RUN ["git", "clone", "https://github.com/bats-core/bats-assert"]

# Copy configure_minio but don't run it yet since it's possible only after minio launch
# (which is made on start)
WORKDIR $HOME
COPY ["tests/configure_minio.sh", "."]

# Copy files and build multisig tool itself
WORKDIR $HOME
RUN ["mkdir", "multisig"]
COPY [".", "multisig"]
WORKDIR multisig
RUN ["go", "install"]

WORKDIR $HOME
COPY ["tests/config_1.toml", "./user1/config.toml"]
COPY ["tests/config_2.toml", "./user2/config.toml"]
COPY ["tests/test_positive.sh", "."]

WORKDIR $HOME
COPY ["tests/start.sh", "."]
ENTRYPOINT ["/bin/bash"]
CMD ["start.sh"]
