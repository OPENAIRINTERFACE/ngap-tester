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

import logging
import subprocess
import sys
import time

logging.basicConfig(
    level=logging.DEBUG,
    stream=sys.stdout,
    format="[%(asctime)s] %(name)s:%(levelname)s: %(message)s"
)

NB_PROFILES = 3

def main() -> None:
    logging.debug('\033[0;32m OMEC gnbsim RAN emulator started, checking if all profiles finished... takes few secs\033[0m....')
    # First using docker ps to see which images were used.
    cmd = 'docker ps -a'
    res = run_cmd(cmd, False)
    print(res)
    notSilentForFirstTime = False
    status = 0
    for x in range(40):
        cmd = f'docker logs omec-gnbsim 2>&1 | grep --colour=never "Profile " || true'
        res = run_cmd(cmd, notSilentForFirstTime)
        notSilentForFirstTime = True
        if res is None:
            exit(f'\033[0;31m Incorrect/Unsupported executing command "{cmd}"')
        cnt = res.count('Profile Status:')
        passing = res.count('Profile Status: PASS')
        if cnt == NB_PROFILES:
            logging.debug('\033[0;32m All profiles finished\033[0m....')
            if passing == NB_PROFILES:
                logging.debug('\033[0;32m All profiles passed\033[0m....')
            else:
                logging.error('\033[0;32m Some profiles failed\033[0m....')
                status = -1
            print(res)
            break
        time.sleep(2)
    if cnt != NB_PROFILES:
        logging.error('\033[0;31m Some profiles could not finish\033[0m....')
        print(res)
        sys.exit(-1)
    sys.exit(status)

def run_cmd(cmd, silent=True):
    if not silent:
        logging.debug(cmd)
    result = None
    try:
        res = subprocess.run(cmd,
                        shell=True, check=True,
                        stdout=subprocess.PIPE,
                        universal_newlines=True)
        result = res.stdout.strip()
    except:
        pass
    return result

if __name__ == '__main__':
    main()
