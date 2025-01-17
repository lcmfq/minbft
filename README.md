[![GitHub Actions](https://github.com/hyperledger-labs/minbft/workflows/Continuous%20integration/badge.svg)](https://github.com/hyperledger-labs/minbft/actions)
[![GoDoc](https://godoc.org/github.com/hyperledger-labs/minbft?status.svg)](https://godoc.org/github.com/hyperledger-labs/minbft)
[![Go Report Card](https://goreportcard.com/badge/github.com/hyperledger-labs/minbft)](https://goreportcard.com/report/github.com/hyperledger-labs/minbft)
[![Total alerts](https://img.shields.io/lgtm/alerts/g/hyperledger-labs/minbft.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/hyperledger-labs/minbft/alerts/)

# MinBFT #

  * [Status](#status)
  * [What is MinBFT](#what-is-minbft)
  * [Why MinBFT](#why-minbft)
  * [Concepts](#concepts)
  * [Quick Start](#quick-start)
      * [Prerequisites](#prerequisites)
      * [Building a Container Image](#building-a-container-image)
      * [Running Replicas](#running-replicas)
      * [Submitting Requests](#submitting-requests)
      * [Tear Down](#tear-down)
  * [Requirements](#requirements)
      * [Operating System](#operating-system)
      * [Golang](#golang)
      * [Intel® SGX SDK](#intel-sgx-sdk)
  * [Getting Started](#getting-started)
      * [Building](#building)
      * [Running Example](#running-example)
      * [Fault Tolerance](#fault-tolerance)
      * [Code Structure](#code-structure)
  * [Roadmap](#roadmap)
  * [Contributing](#contributing)
  * [License](#license)

## Status ##

This project is in experimental development stage. It is **not**
suitable for any kind of production use. Interfaces and implementation
may change significantly during the life cycle of the project.

## What is MinBFT ##

MinBFT is a pluggable software component that allows to achieve
Byzantine fault-tolerant consensus with fewer consenting nodes and
less communication rounds comparing to the conventional BFT protocols.
The component is implemented based
on [Efficient Byzantine Fault-Tolerance paper][minbft-paper], a BFT
protocol that leverages secure hardware capabilities of the
participant nodes.

This project is an implementation of MinBFT consensus protocol. The
code is written in Golang, except the TEE part which is written in C,
as an Intel® SGX enclave.

[minbft-paper]: https://www.researchgate.net/publication/260585535_Efficient_Byzantine_Fault-Tolerance

## Why MinBFT ##

[Byzantine fault-tolerant][bft-wiki] (BFT) protocols are able to
achieve high transaction throughput in permission-based consensus
networks with static configuration of connected nodes. However,
existing components such as [Practical Byzantine Fault Tolerance][pbft-paper]
(PBFT) consensus plugin still incur high communication cost. Given
the pervasive [Trusted Execution Environments][tee-wiki] (TEE) on
commodity platforms, we see a strong motivation to deploy more
efficient BFT protocols that leverage the TEE applications as trust
anchors even on faulty nodes.

TEEs are pervasive nowadays and supported by many commodity platforms.
For instance, Intel® SGX is being deployed on many PCs and servers,
while mobile platforms are mostly supported by Arm Trustzone. TEEs
rely on secure hardware to provide data protection and code isolation
from a compromised hosting system. Here, we propose such a consensus
component that implements MinBFT, a BFT protocol that leverages TEEs
to prevent message equivocation, and thus reducing the required number
of consenting nodes as well as communication rounds given the same
fault-tolerance threshold. More specifically, it requires only `2f+1`
consenting nodes in order to tolerate `f` faulty nodes (or tolerate up
to half of faulty nodes); meanwhile, committing a message requires
only 2 rounds of communication among the nodes instead of 3 rounds as
in PBFT.

We hope that by evaluating this consensus component under the existing
blockchain frameworks, the community will benefit from availability to
leverage it in different practical use cases.

[bft-wiki]: https://en.wikipedia.org/wiki/Byzantine_fault
[pbft-paper]: http://pmg.csail.mit.edu/papers/osdi99.pdf
[tee-wiki]: https://en.wikipedia.org/wiki/Trusted_execution_environment

## Concepts ##

The consensus process in MinBFT is very similar to PBFT. Consenting
nodes (i.e., nodes who vote for ordering the transactions) compose a
fully-connected network. There is a leader (often referred to as
primary node) among the nodes who first prepares an incoming request
message to the other nodes by suggesting a sequence number for the
request in broadcasted `PREPARE` messages. The other nodes verify the
`PREPARE` messages and subsequently broadcast a `COMMIT` message in
the network. Finally, the nodes who have received `f+1` (where `f` is
fault-tolerance, a number of tolerated faulty nodes) consistent
`COMMIT` messages execute the request locally and update the
underlying service state. If the leader is perceived as faulty, a view
change procedure follows to change the leader node.

Note that the signatures for each `PREPARE` and `COMMIT` messages are
generated by USIG (Unique Sequential Identifier Generator) service,
the tamper-proof part of the consenting nodes. The sequence numbers
are also assigned to the messages by USIG with the help of a unique,
monotonic, and sequential counter protected by TEE. The signature,
also known as USIG certificate, certifies the counter assigned to a
particular message. The USIG certificate combined with the counter
value comprises a UI (unique identifier) of the message. Since the
monotonic counter prevents a faulty node from sending conflicting
messages to different nodes and provides correct information during
view changes, MinBFT requires less communication rounds and total
number of consenting nodes than PBFT.

For more detailed description of the protocol, refer to [Efficient
Byzantine Fault-Tolerance paper][minbft-paper].

## Quick Start ##

This quick start shows you how to run this project using Docker
containers. This is the easiest way to try out this project with minimal
setup. If you want to build and run without Docker, skip this section.

### Prerequisites ###

To run the containers, the following software must be installed on your
system.

 - [Docker Engine](https://docs.docker.com/engine/) (tested with version
   19.03.8)
 - [Docker Compose](https://docs.docker.com/compose/) (tested with
   version 1.25.0)

If you are using Ubuntu 20.04, they can be installed as follows:

```sh
sudo apt-get install docker.io docker-compose
```

Note that SGX-enabled CPU is not required to run the containers; they
will run in the simulation mode provided by Intel SGX SDK. We plan to
have another container image which runs in the HW mode (i.e. using
"real" hardware features) in a future release.

### Building a Container Image ###

Build an image of the containers as follows:

```sh
sudo docker build -f sample/docker/Dockerfile -t minbft .
```

Note that, by defaut, the `docker` command needs to be executed by
`root`. Refer to [Docker's
document](https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user)
for details.

If your system has the `make` command, the following command also can be
used.

```sh
sudo make docker
```

### Running Replicas ###

To start up an example consensus network of replicas, invoke the
following commands:

```sh
sudo UID=$UID docker-compose -f sample/docker/docker-compose.yml up -d
```

This will start the replica nodes as 3 separate containers in
background.

### Submitting Requests ###

Requests can be submitted for ordering and execution to the example
consensus network as follows:

```sh
sudo docker-compose -f sample/docker/docker-compose.yml \
  run client request "First request" "Second request" "Another request"
```

This command should produce the following output showing the result
of ordering and execution of the submitted requests:

```
Reply: {"Height":1,"PrevBlockHash":null,"Payload":"Rmlyc3QgcmVxdWVzdA=="}
Reply: {"Height":2,"PrevBlockHash":"DuAGbE1hVQCvgi+R0E5zWaKSlVYFEo3CjlRj9Eik5h4=","Payload":"U2Vjb25kIHJlcXVlc3Q="}
Reply: {"Height":3,"PrevBlockHash":"963Kn659GbtX35MZYzguEwSH1UvF2cRYo6lNpIyuCUE=","Payload":"QW5vdGhlciByZXF1ZXN0"}
```

The output shows the submitted requests being ordered and executed by a
sample blockchain service. The service executes request by simply
appending a new block for each request to the trivial blockchain
maintained by the service.

### Tear Down ###

The following command can be used to terminate running containers:

```sh
sudo docker-compose -f sample/docker/docker-compose.yml down
```

The containers create some files while running. These files can be
deleted as follows.

```sh
rm -f sample/docker/keys.yaml sample/docker/.keys.yaml.lock
```

The Docker image can be deleted as follow.

```sh
sudo docker rmi minbft
```

If your system has the `make` command, the following command also can be
used.

```sh
sudo make docker-clean
```

## Requirements ##

### Operating System ###

The project has been tested on Ubuntu 18.04 and 20.04 LTS.
Additional required packages can be installed as follows:

```sh
sudo apt-get install build-essential pkg-config
```

### Golang ###

Go 1.13 or later is required to build this project (tested against
1.13, 1.15 and 1.16). Official installation instructions can be found
[here][go-install].

[go-install]: https://golang.org/doc/install

### Intel® SGX SDK ###

The Intel® SGX enclave implementation has been tested with Intel® SGX
SDK for Linux version 2.12. For installation instuctions please visit
[download page][sgx-downloads].
Please note that Intel SGX has two operation modes and required software
components depend on operation mode.

 - If you run in HW mode, you have to install all three components:
   SGX driver, PSW, and SGX SDK.
 - If you run in simulation mode, only SGX SDK is required.

A conventional directory to install
the SDK is `/opt/intel/`. Please do not forget to source
`/opt/intel/sgxsdk/environment` file in your shell. Alternatively, one
can add the following line to `~/.profile`:

```sh
. /opt/intel/sgxsdk/environment
```

If you run in simlation mode, you need create/update the link to
the additional directory of shared libraries with following commands:

```
sudo bash -c "echo /opt/intel/sgxsdk/sdk_libs > /etc/ld.so.conf.d/sgx-sdk.conf"
sudo ldconfig
```

When using a machine with no SGX support, only SGX simulation mode is
supported. In that case, please be sure to export the following
environment variable, e.g. by modifying `~/.profile` file:

```sh
export SGX_MODE=SIM
```

[sgx-downloads]: https://01.org/intel-software-guard-extensions/downloads

## Getting Started ##

This is a Go module and can be placed anywhere; no need to be in
GOPATH. If this is placed in GOPATH and you are using Go 1.11 or 1.12,
please make sure that the environment variable `GO111MODULE=on` has
set to activate module mode.

All following commands are supposed to be run in the root of the
module's source tree.

### Building ###

The project can be build by issuing the following command. At the
moment, the binaries are installed in `sample/bin/` directory; no root
privileges are needed.

```sh
make install
```

### Running Example ###

Running the example requires some set up. Please make sure the project
has been successfully built and `sample/bin/keytool` and
`sample/bin/peer` binaries were produced. Those binaries can be
supplied with options through a configuration file, environment
variables, or command-line arguments. More information about available
options can be queried by invoking those binaries with `help`
argument. Sample configuration files can be found in
`sample/authentication/keytool/` and `sample/peer/` directories
respectively.

Before running the example, the environment variable `$LD_LIBRARY_PATH`
needs to include `sample/lib` where `libusig_shim.so` is installed by
`make install`.

```sh
export LD_LIBRARY_PATH="${PWD}/sample/lib:${LD_LIBRARY_PATH}"
```

#### Generating Keys ####

The following command are to be run from `sample` directory.

```sh
cd sample
```

Sample key set file for testing can be generate by using `keytool`
command. This command produces a key set file suitable for running the
example on a local machine:

```sh
bin/keytool generate -u lib/libusig.signed.so
```

This invocation will create a sample key set file named `keys.yaml`
containing 3 key pairs for replicas and 1 key pair for a client by
default.

#### Consensus Options Configuration ####

Consensus options can be set up by means of a configuration file. A
sample consensus configuration file can be used as an example:

```sh
cp config/consensus.yaml ./
```

#### Peer Configuration ####

Peer configuration can be supplied in a configuration file. Selected
options can be modified through command line arguments of `peer`
binary. A sample configuration can be used as an example:

```sh
cp peer/peer.yaml ./
```

#### Running Replicas ####

To start up an example consensus network of replicas on a local
machine, invoke the following commands:

```sh
bin/peer run 0 &
bin/peer run 1 &
bin/peer run 2 &
```

This will start the replica nodes as 3 separate OS processes in
background using the configuration files prepared in previous steps.

#### Submitting Requests ####

Requests can be submitted for ordering and execution to the example
consensus network using the same `peer` binary and configuration files
for convenience. It is better to issue the following commands in
another terminal so that the output messages do not intermix:

```sh
bin/peer request "First request" "Second request" "Another request"
```

This command should produce the following output showing the result
of ordering and execution of the submitted requests:

```
Reply: {"Height":1,"PrevBlockHash":null,"Payload":"Rmlyc3QgcmVxdWVzdA=="}
Reply: {"Height":2,"PrevBlockHash":"DuAGbE1hVQCvgi+R0E5zWaKSlVYFEo3CjlRj9Eik5h4=","Payload":"U2Vjb25kIHJlcXVlc3Q="}
Reply: {"Height":3,"PrevBlockHash":"963Kn659GbtX35MZYzguEwSH1UvF2cRYo6lNpIyuCUE=","Payload":"QW5vdGhlciByZXF1ZXN0"}
```

The output shows the submitted requests being ordered and executed by
a sample blockchain service. The service executes request by simply
appending a new block for each request to the trivial blockchain
maintained by the service.

#### Tear Down ####

The following command can be used to terminate running replica
processes and release the occupied TCP ports:

```sh
killall peer
```

### Fault Tolerance ###

The above example shows a simple normal case of the consensus network.
Our next interest is how the system behaves when some replicas are faulty.

#### Crash Fault on Backup ####

The simplest faulty case is crash fault on backup replicas. Note that
we don't tolerate any type of fault on primary replica until view change
operation is implemented.

Let's restart the network, and note the process IDs of each replica process.

```sh
$ bin/peer run 0 &
[1] 16899
$ bin/peer run 1 &
[2] 16916
$ bin/peer run 2 &
[3] 16923
```

Make sure that all replicas are properly working by sending a request:

```sh
$ bin/peer request First request
Reply: {"Height":1,"PrevBlockHash":null,"Payload":"Rmlyc3QgcmVxdWVzdA=="}
```

Now kill replica 1 and send another request:

```sh
$ kill 16916
$ bin/peer request Second request
Reply: {"Height":2,"PrevBlockHash":"DuAGbE1hVQCvgi+R0E5zWaKSlVYFEo3CjlRj9Eik5h4=","Payload":"U2Vjb25kIHJlcXVlc3Q="}
```

OK, we still get the reply messages with successfully agreed response.
Next, kill another backup replica and send another request:

```sh
$ kill 16923
$ bin/peer request Another request
(no response)
```

We fail to reach consensus and get no response because more than
`f` replicas are faulty.

### Code Structure ###

The code divided into core consensus protocol implementation and
sample implementation of external components required to interact with
the core. The following directories contain the code:

  * `api/` - definition of API between core and external components
  * `client/` - implementation of client-side part of the protocol
  * `core/` - implementation of core consensus protocol
  * `usig/` - implementation of USIG, tamper-proof component
  * `messages/` - definition of the protocol messages
  * `sample/` - sample implementation of external interfaces
    * `authentication/` - generation and verification of
                          authentication tags
      * `keytool/` - tool to generate sample key set file
    * `conn/` - network connectivity
    * `config/` - consensus configuration provider
    * `requestconsumer/` - service executing ordered requests
    * `peer/` - CLI application to run a replica/client instance

## Roadmap ##

The following features of MinBFT protocol has been implemented:

  * _Normal case operation_: minimal ordering and execution of
    requests as long as primary replica is not faulty
  * _SGX USIG_: implementation of USIG service as Intel® SGX enclave

The following features are considered to be implemented:

  * _View change operation_: provide liveness in case of faulty
    primary replica
  * _Garbage collection and checkpoints_: generation and handling of
    `CHECKPOINT` messages, log pruning, high and low water marks
  * _USIG enclave attestation_: support to remotely attest USIG
    Intel® SGX enclave
  * _Faulty node recovery_: support to retrieve missing log entries
    and synchronize service state from other replicas
  * _Request batching_: reducing latency and increasing throughput by
    combining outstanding requests for later processing
  * _Asynchronous requests_: enabling parallel processing of requests
  * _MAC authentication_: using MAC in place of digital signature in
    USIG to reduce message size
  * _Read-only requests_: optimized processing of read-only requests
  * _Speculative request execution_: reducing processing delay by
    tentatively executing requests
  * _Documentation improvement_: comprehensive documentation
  * _Testing improvement_: comprehensive unit- and integration tests
  * _Benchmarks_: measuring performance

## Contributing ##

Everyone is welcome to contribute! There are many ways to make useful
contribution. Please look at [Contribution Guideline](CONTRIBUTING.md)
for more details.

## License ##

Source code files are licensed under
the [Apache License, Version 2.0](LICENSE).

Documentation files are licensed under
the [Creative Commons Attribution 4.0 International License][cc-40].

[cc-40]: http://creativecommons.org/licenses/by/4.0/
