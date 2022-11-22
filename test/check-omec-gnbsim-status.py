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

NB_PROFILES = 4
NB_PROFILES_2 = 1

def main() -> None:
    logging.debug('\033[0;32m OMEC gnbsim RAN emulator started, checking if all profiles finished... takes few secs\033[0m....')
    # First using docker ps to see which images were used.
    cmd = 'docker ps -a'
    res = run_cmd(cmd, False)
    print(res)
    notSilentForFirstTime = False
    status = 0
    for x in range(100):
        cmd1 = f'docker logs omec-gnbsim-1 2>&1 | grep --colour=never "Profile " | grep -v "Waiting for UEs to finish processing" || true'
        res1 = run_cmd(cmd1, notSilentForFirstTime)
        cmd2 = f'docker logs omec-gnbsim-2 2>&1 | grep --colour=never "Profile " | grep -v "Waiting for UEs to finish processing" || true'
        res2 = run_cmd(cmd2, notSilentForFirstTime)
        cmd3 = f'docker logs omec-gnbsim-3 2>&1 | grep --colour=never "Profile " | grep -v "Waiting for UEs to finish processing" || true'
        res3 = run_cmd(cmd3, notSilentForFirstTime)
        cmd4 = f'docker logs omec-gnbsim-4 2>&1 | grep --colour=never "Profile " | grep -v "Waiting for UEs to finish processing" || true'
        res4 = run_cmd(cmd4, notSilentForFirstTime)
        notSilentForFirstTime = True
        if res1 is None or res2 is None or res3 is None or res4 is None:
            exit(f'\033[0;31m Incorrect/Unsupported executing command "{cmd}"')
        cnt1 = res1.count('Profile Status:')
        cnt2 = res2.count('Profile Status:')
        cnt3 = res3.count('Profile Status:')
        cnt4 = res4.count('Profile Status:')
        passing1 = res1.count('Profile Status: PASS')
        passing2 = res2.count('Profile Status: PASS')
        passing3 = res3.count('Profile Status: PASS')
        passing4 = res4.count('Profile Status: PASS')
        if cnt1 == NB_PROFILES and cnt2 == NB_PROFILES_2 and cnt3 == NB_PROFILES_2 and cnt4 == NB_PROFILES_2:
            logging.debug('\033[0;32m All profiles finished\033[0m....')
            if passing1 == NB_PROFILES and passing2 == NB_PROFILES_2 and passing3 == NB_PROFILES_2 and passing4 == NB_PROFILES_2:
                logging.debug('\033[0;32m All profiles passed\033[0m....')
            else:
                logging.error('\033[0;32m Some profiles failed\033[0m....')
                status = -1
            print(res1)
            print(res2)
            print(res3)
            print(res4)
            break
        time.sleep(10)
    cmd = 'docker ps -a'
    res = run_cmd(cmd, False)
    print(res)
    if cnt1 != NB_PROFILES or cnt2 != NB_PROFILES_2 or cnt3 != NB_PROFILES_2 or cnt4 != NB_PROFILES_2:
        logging.error('\033[0;31m Some profiles could not finish\033[0m....')
        print(res1)
        print(res2)
        print(res3)
        print(res4)
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
