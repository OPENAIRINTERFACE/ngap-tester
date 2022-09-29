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

NB_HEALTHY_CONTAINERS_IN_BASIC_NRF = 8

def main() -> None:
    logging.debug('\033[0;32m OAI 5G Core network started, checking the health status of the containers... takes few secs\033[0m....')
    notSilentForFirstTime = False
    for x in range(40):
        cmd = f'docker-compose -f docker-compose-basic-nrf.yaml ps -a'
        res = run_cmd(cmd, notSilentForFirstTime)
        notSilentForFirstTime = True
        if res is None:
            exit(f'\033[0;31m Incorrect/Unsupported executing command "{cmd}"')
        cnt = res.count('(healthy)')
        if cnt == NB_HEALTHY_CONTAINERS_IN_BASIC_NRF:
            logging.debug('\033[0;32m All components are healthy, please see below for more details\033[0m....')
            print(res)
            break
        time.sleep(2)
    if cnt != NB_HEALTHY_CONTAINERS_IN_BASIC_NRF:
        logging.error('\033[0;31m Core network is un-healthy, please see below for more details\033[0m....')
        print(res)
        sys.exit(-1)
    sys.exit(0)

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
