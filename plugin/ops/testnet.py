from os import path, mkdir
from typing import Tuple, List
from subprocess import run, check_output, PIPE, STDOUT
from socket import socket
from tempfile import gettempdir
from shutil import rmtree

from config import setup_config
from utils import generate_service, check_docker_is_installed

CONFIG = setup_config()
REPOS = CONFIG["REPOS"]

clientTomlTemplate = """
[Logging]
  Disable = false
  Level = "DEBUG"
  File = ""

[UpstreamProxy]
  Type = "none"

[Debug]
  DisableDecoyTraffic = {}
  CaseSensitiveUserIdentifiers = false
  PollingInterval = 1

[NonvotingAuthority]
    Address = "{}:{}"
    PublicKey = "{}"
"""

def generate_mixnet_config(ip: str, confDir: str) -> List[str]:
    """
    Runs the genconfig tool at a specified directory
    and returns a list of paths of the generate configs
    """
    try: rmtree(confDir)
    except FileNotFoundError: pass
    mkdir(confDir)
    output = run([
        "genconfig",
        "-o",
        confDir,
        "-n",
        str(CONFIG["TEST"]["NODES"]),
        "-a",
        ip,
        "-p",
        str(CONFIG["TEST"]["PROVIDERS"]),
    ], stdout=PIPE, stderr=STDOUT)
    return [p.split(" ")[-1] for p in output.stdout.decode().strip().split("\n")]

def get_public_key(dirPath: str) -> str:
    """Gets the public key from identity.public.pem file in given directory"""
    with open(path.join(dirPath, "identity.public.pem"), 'r') as f:
        return f.read().split("\n")[1] # line index 1


def get_mixnet_port(path: str) -> str:
    """Gets mixnet port number from a given katzenpost.toml file"""
    with open(path, 'r') as f:
        for line in f:
            if "Addresses = [" in line:
                return line.split('"')[1].split(":")[1]

def get_user_registration_port(path: str) -> str:
    """Gets the user registration port from a given katzenpost.toml file"""
    with open(path, 'r') as f:
        for line in f:
            if "UserRegistrationHTTPAddresses" in line:
                return line.split('"')[1].split(":")[1]

def get_data_dir(path: str) -> str:
    """Gets the config directory path from the given config file"""
    with open(path, 'r') as f:
        for line in f:
            if "DataDir =" in line:
                return line.split('=')[-1].replace('"', '').strip()

def get_ip() -> str:
    """Gets the IP address that is accesible by all containers"""
    s = socket()
    s.connect(("1.1.1.1", 80))
    ip = s.getsockname()[0]
    s.close()
    return ip

def run_docker(ip: str, composePath: str) -> None:
    """Starts Docker stack deploy. Starts a docker swarm if there isn't one"""
    output = check_output(["docker", "info"])
    if "Swarm: inactive" in output.decode():
        run(["docker", "swarm", "init", "--advertise-addr={}".format(ip)], check=True)

    args = ["docker", "stack", "deploy", "-c", composePath, "mixnet"]
    try:
        run(args, check=True)
    except:
        log("Failed in deploying docker compose", True)

    run(["docker", "service",  "ls", "--format", '"{{.Name}} with tag {{.Image}}"'])

def main():
    check_docker_is_installed()
    testnetConfDir = path.join(gettempdir(), "meson-testnet")
    ip = get_ip()
    confPaths = generate_mixnet_config(ip, testnetConfDir)

    authToml = [p for p in confPaths if "nonvoting" in p][0]
    confPaths.remove(authToml)
    # Save client.toml
    with open(path.join(testnetConfDir, "client.toml"), 'w+') as f:
        f.write(clientTomlTemplate.format(
            "true",
            ip,
            get_mixnet_port(authToml),
            get_public_key(path.join(path.dirname(authToml)))
        ))

    authorityYAML = generate_service(
        name="authority",
        image=REPOS["AUTH"]["CONTAINER"]+":"+REPOS["AUTH"]["NAMEDTAG"],
        ports=[
            "{0}:{0}".format(get_mixnet_port(authToml))
        ],
        volumes=[
            "{}:{}".format(path.join(path.dirname(authToml)), get_data_dir(authToml))
        ]
    )

    # We set this value with 35000 because there is no config file 
    # that we can scrape that has this value.
    currentPrometheusPort = 35000
    containerYAML = ""
    for toml in confPaths:
        currentPrometheusPort += 1
        confDir = path.dirname(toml)
        name = path.basename(confDir)
        ports = [
            "{0}:{0}".format(get_mixnet_port(toml)),
            "{}:{}".format(currentPrometheusPort, "6543"),
        ]
        if get_user_registration_port(toml):
            ports.append("{0}:{0}".format(get_user_registration_port(toml)))

        containerYAML += generate_service(
            name=name,
            image=REPOS["MESON"]["CONTAINER"]+":"+REPOS["MESON"]["NAMEDTAG"],
            ports=ports,
            volumes=[
                "{}:{}".format(confDir, get_data_dir(toml))
            ],
            dependsOn=["authority"]
        )

    # save compose file
    composePath = path.join(testnetConfDir, "testnet.yml")
    with open(composePath, 'w+') as f:
        f.write('version: "3.7"\nservices:\n' + authorityYAML + containerYAML)

    run_docker(ip, composePath)

if __name__ == "__main__":
    main()
