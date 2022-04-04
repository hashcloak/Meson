from subprocess import run

from config import setup_config
from utils import log

CONFIG = setup_config()

def main():
    for repo in CONFIG["REPOS"].values():
        log("Pushing container {}:{}".format(repo["CONTAINER"], repo["NAMEDTAG"]))
        run(["docker", "push", "{}:{}".format(repo["CONTAINER"], repo["NAMEDTAG"])])
        log("Pushing container {}:{}".format(repo["CONTAINER"], repo["HASHTAG"]))
        run(["docker", "push", "{}:{}".format(repo["CONTAINER"], repo["HASHTAG"])])

if __name__ == "__main__":
    main()
