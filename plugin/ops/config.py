from os import getenv
from typing import List, Tuple
from subprocess import check_output

CONFIG = {
    "REPOS": {
        "AUTH": {
            "CONTAINER": "katzenmint/pki",
            "REPOSITORY": "https://github.com/hashcloak/Meson",
            "SUBDIRECTORY": "katzenmint",
            "BRANCH": "monorepo",
            "GITHASH": "",
            "NAMEDTAG": "latest",
            "HASHTAG": "",
        },
        "SERVER" : {
            "CONTAINER": "meson/server",
            "REPOSITORY": "https://github.com/hashcloak/Meson",
            "SUBDIRECTORY": "server",
            "BRANCH": "monorepo",
            "GITHASH": "",
            "NAMEDTAG": "latest",
            "HASHTAG": "",
        },
    },
    "TEST": {
        "PKS": {
            "ETHEREUM": "",
            "BINANCE": ""
        },
        "CLIENTCOMMIT": "master",
        "NODES": 2,
        "PROVIDERS": 2,
        "ATTEMPTS": 3,
    },
    "LOG": "",
    "WARPED": True,
    "WITHOUTCACHE": True,
    "BUILD": "",
}

HASH_LENGTH=7
def get_remote_git_hash(repositoryURL: str, branchOrTag: str) -> str:
    """Gets the first 7 characters of a git commit hash in a remote repository"""
    args = ["git", "ls-remote", repositoryURL, branchOrTag]
    return check_output(args).decode().split('\t')[0][:HASH_LENGTH]

def get_local_repo_info() -> Tuple[str, str]:
    """
    Gets the local repository information.
    This is changes depending on whether it is is running in Travis.
    """
    arguments = ["git", "rev-parse", "--abbrev-ref", "HEAD"]
    gitBranch = check_output(arguments).decode().strip()
    arguments = ["git", "rev-parse", "HEAD"]
    gitHash = check_output(arguments).decode().strip()
    if getenv('TRAVIS_EVENT_TYPE') == "pull_request":
        gitBranch = getenv('TRAVIS_PULL_REQUEST_BRANCH', gitBranch)
        gitHash = getenv('TRAVIS_PULL_REQUEST_SHA', gitHash)
    else:
        gitBranch = getenv('TRAVIS_BRANCH', gitBranch)
        gitHash = getenv('TRAVIS_COMMIT', gitHash)

    return gitBranch, gitHash[:HASH_LENGTH]

def expand_dict(dictionary: dict, separator="_") -> List[str]:
    """
    Joins all the keys of a dictionary with a separator string
    separator default is '_'
    """
    tempList = []
    for key, value in dictionary.items():
        if type(value) == dict:
            tempList.extend([key+separator+item for item in expand_dict(value)])
        else:
            tempList.append(key)

    return tempList

def set_nested_value(dictionary: dict, value: str, keys: List[str]) -> None:
    """Sets a nested value inside a dictionary"""
    if keys and dictionary:
        if len(keys) == 1:
            dictionary[keys[0]] = value
        else:
            set_nested_value(dictionary.get(keys[0]), value, keys[1:])

def get_nested_value(dictionary: dict, *args: List[str]) -> str:
    """Gets a nested value from a dictionary"""
    if args and dictionary:
        subkey = args[0]
        if subkey:
            value = dictionary.get(subkey)
            return value if len(args) == 1 else get_nested_value(value, *args[1:])

def setup_config() -> dict:
    for envVar in expand_dict(CONFIG):
        value = getenv(envVar, get_nested_value(CONFIG, *envVar.split("_")))
        set_nested_value(CONFIG, value, envVar.split("_"))

    localBranch, localHash = get_local_repo_info()

    for key, repo in CONFIG["REPOS"].items():
        hashValue = localHash
        repo["GITHASH"] = repo["GITHASH"] if repo["GITHASH"] else hashValue
        repo["HASHTAG"] = "warped_"+repo["GITHASH"] if CONFIG["WARPED"] else repo["GITHASH"]
        if len(repo["NAMEDTAG"]) == 0:
            repo["NAMEDTAG"] = "warped_"+repo["BRANCH"] if CONFIG["WARPED"] else repo["BRANCH"]

    return CONFIG
