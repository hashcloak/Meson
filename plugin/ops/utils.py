from config import setup_config
from subprocess import check_output, run, PIPE, STDOUT
from typing import List
import sys

CONFIG = setup_config()
def log(message: str, err=False, forceLog=False) -> None:
    """Logs a message to the console with purple.
    If err is True then it will log with red"""
    color = '\033[0;31m' if err else '\033[0;32m'
    noColor='\033[0m' # No Color
    if CONFIG["LOG"] or forceLog:
        print("{}LOG: {}{}".format(color, message, noColor))

def check_docker_is_installed() -> None:
    try:
        check_output(["docker", "info"])
    except:
        print("Docker not found. Please install docker")
        sys.exit(1)

def checkout_repo(repoPath: str, repoUrl: str, commitOrBranch: str) -> None:
    """Clones, and git checkouts a repository given a path, repo url and a commit or branch"""
    output = run(["git", "clone", repoUrl, repoPath], stdout=PIPE, stderr=STDOUT)
    safeError = 'already exists and is not an empty directory' in output.stdout.decode()
    log(output.stdout.decode().strip(), not safeError)
    if safeError:
        log("Ignoring last error, continuing...")

    run(["git", "fetch"], check=True, cwd=repoPath)
    run(["git", "reset", "--hard"], check=True, cwd=repoPath)
    run(["git", "-c", "advice.detachedHead=false", "checkout", commitOrBranch], check=True, cwd=repoPath)

def generate_service(
    name: str,
    image: str,
    ports: List[str] = [],
    volumes: List[str] = [],
    dependsOn: List[str] = [],
) -> str:
    """
    Creates a string with docker compose service specification.
    Arguments are a list of values that need to be added to each section
    named after the parameter. i.e. the volume arguments are for the
    volumes section of the service config.
    """
    indent = '  '
    service = "{s}{name}:\n{s}{s}image: {image}\n".format(
        s=indent,
        name=name,
        image=image,
    )

    if ports:
        service += "{s}ports:\n".format(s=indent*2)
        for port in ports:
            service += '{s}- "{port}"\n'.format(s=indent*3, port=port)

    if volumes:
        service += "{s}volumes:\n".format(s=indent*2)
        for vol in volumes:
            service += '{s}- {vol}\n'.format(s=indent*3, vol=vol)

    if dependsOn:
        service += "{s}depends_on:\n".format(s=indent*2)
        for item in dependsOn:
            service += '{s}- "{dep}"\n'.format(s=indent*3, dep=item)

    return service
