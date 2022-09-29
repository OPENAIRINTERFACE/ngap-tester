#!/usr/bin/env python3

"""
Copyright 2020 The Magma Authors.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
"""

import argparse
import os
import re
import subprocess  # noqa: S404
import sys

DEFAULT_GNBSIM_BRANCH='ngap-tester'
DEFAULT_NAS_BRANCH='ngap-tester'
DEFAULT_NGAP_BRANCH='ngap-tester'

def main() -> None:
    """Provide command-line options to check/synchronize git sub-modules"""
    args = _parse_args()

    if args.check and not args.synchronize:
        status_check()
        sys.exit(0)

    if args.synchronize:
        ret = status_check()
        if not ret and not args.force:
            str_input = input('Some files are modified and/or untracked. Do you want to continue (y/N)?')
            if str_input == 'y' or str_input == 'Y':
                print('We continue')
            elif str_input == '' or str_input == 'n' or str_input == 'N':
                print('We stop')
                sys.exit(0)
            else:
                sys.exit("I don't understand your answer. Stopping")
        print('Synchronizing now')
        synchronize(args)
        sys.exit(0)

    sys.exit('No option specified')

def _parse_args() -> argparse.Namespace:
    """Parse the command line args

    Returns:
        argparse.Namespace: the created parser
    """
    parser = argparse.ArgumentParser(description='Script to check/synchronize git sub-modules')

    parser.add_argument(
        '--check', '-ck',
        action='store_true',
        default=False,
        help='Only performs a check on whole repo and its sub-modules in this workspace',
    )

    parser.add_argument(
        '--synchronize', '-sy',
        action='store_true',
        default=False,
        help='Will synchronize git submodules to specified/default branches/commits',
    )

    parser.add_argument(
        '--gnbsim',
        action='store',
        default=DEFAULT_GNBSIM_BRANCH,
        help='Specify the gnbsim sub-module branch / commit to synchronize with',
    )

    parser.add_argument(
        '--nas',
        action='store',
        default=DEFAULT_NAS_BRANCH,
        help='Specify the gnbsim sub-module branch / commit to synchronize with',
    )

    parser.add_argument(
        '--ngap',
        action='store',
        default=DEFAULT_NGAP_BRANCH,
        help='Specify the gnbsim sub-module branch / commit to synchronize with',
    )

    parser.add_argument(
        '--force', '-f',
        action='store_true',
        default=False,
        help='Will force synchronization! Caution can be dangerous',
    )

    return parser.parse_args()

def git_status_check(response):
    """Parse standard-output for modified and untracked files

    Args:
        response: the 'git status' standard-output
    Returns:
        Boolean: True if none; False if any
    """
    status = True
    for line in response.split('\n'):
        print(line)
        if re.search('Changes to be committed', line):
            status = False
        if re.search('Changes not staged for commit', line):
            status = False
        if re.search('modified:', line):
            status = False
        if re.search('Untracked files:', line):
            status = False
    return status

def status_check():
    """Parse each sub-module for modified and untracked files

    Returns:
        Boolean: True if none; False if any
    """
    cmd = 'git status --ignored'
    print ('Checking main repo --> ' + cmd)
    ret = subprocess.check_output(cmd, shell=True, universal_newlines=True)  # noqa: S602
    status0 = git_status_check(ret)
    cmd = 'git submodule foreach "git status --ignored"'
    print ('Checking sub-modules --> ' + cmd)
    ret = subprocess.check_output(cmd, shell=True, universal_newlines=True)  # noqa: S602
    status1 = git_status_check(ret)
    return (status0 and status1)

def synchronize(args):
    """In each submodule, cleaning and then setting to specified branch/commit

    Args:
        args: the created parser
    Returns:
        Boolean: True if none; False if any
    """
    paths = []
    cmd = 'git submodule status --recursive'
    ret = subprocess.check_output(cmd, shell=True, universal_newlines=True)  # noqa: S602
    for line in ret.split('\n'):
        re_res = re.search('third-party/([A-Za-z0-9\-]+)', line)
        if re_res is not None:
            paths.append('third-party/' + re_res.group(1))

    for sm_path in paths:
        branch = ''
        if sm_path == 'third-party/gnbsim':
            branch = args.gnbsim
        if sm_path == 'third-party/nas':
            branch = args.nas
        if sm_path == 'third-party/ngap':
            branch = args.ngap
        if branch == '':
            continue
        print('Synchronizing ' + sm_path + ' with branch/commit ' + str(branch))
        if not os.listdir('./' + sm_path):
            cmd = 'git submodule init'
            subprocess.check_call(cmd, shell=True, universal_newlines=True)  # noqa: S602
            cmd = 'git submodule update'
            subprocess.check_call(cmd, shell=True, universal_newlines=True)  # noqa: S602
        prefix = 'cd ' + sm_path + ' && '
        cmd = prefix +  'git fetch --prune'
        subprocess.check_call(cmd, shell=True, universal_newlines=True)  # noqa: S602
        cmd = prefix +  'git clean -x -d -ff'
        subprocess.check_call(cmd, shell=True, universal_newlines=True)  # noqa: S602
        cmd = prefix +  'git reset .'
        subprocess.check_call(cmd, shell=True, universal_newlines=True)  # noqa: S602
        cmd = prefix +  'git checkout .'
        subprocess.check_call(cmd, shell=True, universal_newlines=True)  # noqa: S602
        cmd = prefix +  'git checkout ' + branch
        subprocess.check_call(cmd, shell=True, universal_newlines=True)  # noqa: S602
        cmd = prefix +  'git log -n1'
        subprocess.check_call(cmd, shell=True, universal_newlines=True)  # noqa: S602

if __name__ == '__main__':
    main()
