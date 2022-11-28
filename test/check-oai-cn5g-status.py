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
import logging
import subprocess
import sys
import time

logging.basicConfig(
    level=logging.INFO,
    stream=sys.stdout,
    format="[%(asctime)s] %(levelname)8s: %(message)s"
)

NB_HEALTHY_CONTAINERS_IN_BASIC_NRF = 8

def main() -> None:
    args = _parse_args()
    logging.info('\033[0;32m OAI 5G Core network started, checking the health status of the containers... takes few secs\033[0m....')
    if args.vpp_upf:
        logging.info('Using VPP-UPF variant')
        dc_file = 'docker-compose-basic-vpp-nrf.yaml'
    else:
        logging.info('Using SPGWU-TINY variant')
        dc_file = 'docker-compose-basic-nrf.yaml'
    notSilentForFirstTime = False
    for x in range(40):
        cmd = f'docker-compose -f {dc_file} ps -a'
        res = run_cmd(cmd, notSilentForFirstTime)
        notSilentForFirstTime = True
        if res is None:
            exit(f'\033[0;31m Incorrect/Unsupported executing command "{cmd}"')
        cnt = res.count('(healthy)')
        if cnt == NB_HEALTHY_CONTAINERS_IN_BASIC_NRF:
            logging.info('\033[0;32m All components are healthy, please see below for more details\033[0m....')
            print(res)
            break
        time.sleep(2)
    if cnt != NB_HEALTHY_CONTAINERS_IN_BASIC_NRF:
        logging.error('\033[0;31m Core network is un-healthy, please see below for more details\033[0m....')
        print(res)
        sys.exit(-1)
    sys.exit(0)

def _parse_args() -> argparse.Namespace:
    """Parse the command line args

    Returns:
        argparse.Namespace: the created parser
    """
    parser = argparse.ArgumentParser(description='Script to validate if OAI-CN5G is ready and healthy.')

    parser.add_argument(
        '--vpp-upf',
        action='store_true',
        default=False,
        help='Will use vpp-upf variant for deployment if true; spgwu-tiny variant if false (default)',
    )

    return parser.parse_args()

def run_cmd(cmd, silent=True):
    if not silent:
        logging.info(cmd)
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
