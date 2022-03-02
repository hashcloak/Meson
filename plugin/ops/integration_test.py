from os import path, getenv
from sys import exit
from tempfile import gettempdir
from subprocess import run, STDOUT, PIPE
from time import sleep

from config import setup_config
from utils import checkout_repo, log

CONFIG = setup_config()

def main():
    repoPath = path.dirname(path.dirname(path.dirname(path.realpath(__file__))))
    confDir = path.dirname(path.dirname(path.dirname(path.realpath(__file__))))

    warpedBuildFlags='-ldflags "-X github.com/katzenpost/core/epochtime.WarpedEpoch=true -X github.com/katzenpost/server/internal/pki.WarpedEpoch=true"'
    cmd = "go run {warped} {testGo} -c {client} -s ping".format(
        warped=warpedBuildFlags if CONFIG["WARPED"] else "",
        testGo=path.join(repoPath, "ping", "main.go"),
        client=path.join(confDir, "ping", "client.toml"),
    )

    # The attempts are needed until the stability of the mixnet gets improved.
    # This issue is a step in that direction: https://github.com/hashcloak/Meson/plugin/issues/29
    attempts = CONFIG["TEST"]["ATTEMPTS"]
    while True:
        log("Attempt {}: {}".format(attempts, cmd))
        output = run([cmd], stdout=PIPE, stderr=STDOUT, shell=True)
        # Travis has issues printing a huge string.
        # Creating seperate print statements helps with this
        for line in output.stdout.decode().split("\n"):
            log(line, output.returncode == 1)

        if output.returncode == 0:
            log(line, output.returncode)
            exit(0)

        attempts -= 1
        if attempts == 0:
            exit(1)
        sleep(10)

if __name__ == "__main__":
    main()
