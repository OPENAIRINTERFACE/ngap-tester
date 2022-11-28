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
import re
import subprocess
import sys
import time
import matplotlib.pyplot as plt

logging.basicConfig(
    level=logging.INFO,
    stream=sys.stdout,
    format="[%(asctime)s] %(levelname)8s: %(message)s"
)

LOOP_CNT = 60
LOOP_INTERVAL = 5
NB_GNBSIM_INSTANCES = 4
NB_PROFILES = [4, 1, 1, 1]

def main() -> None:
    args = _parse_args()
    if args.vpp_upf:
        NB_PROFILES[0] = 1
    plt.set_loglevel("info")
    logging.info('\033[0;32m OMEC gnbsim RAN emulator started, checking if all profiles finished... takes few secs\033[0m....')
    # First using docker ps to see which images were used.
    cmd = 'docker ps -a'
    res = run_cmd(cmd, False)
    print(res)
    notSilentForFirstTime = False
    status = -1
    # Stats Arrays
    amfTimeX = []
    amfMemY = []
    amfCpuY = []
    nrfTimeX = []
    nrfMemY = []
    nrfCpuY = []
    ausfTimeX = []
    ausfMemY = []
    ausfCpuY = []
    udmTimeX = []
    udmMemY = []
    udmCpuY = []
    udrTimeX = []
    udrMemY = []
    udrCpuY = []
    smfTimeX = []
    smfCpuY = []
    smfMemY = []
    spgwuTimeX = []
    spgwuMemY = []
    spgwuCpuY = []
    for x in range(LOOP_CNT):
        # Performing some statistics measurements on CPU and Memory usages for each NF
        stats = run_cmd('docker stats --no-stream', notSilentForFirstTime)
        for line in stats.split('\n'):
            if line.count('oai-amf') > 0:
                result = re.search(' (?P<cpu_usage>[0-9\.]+)% *(?P<memory_usage>[0-9\.]+)MiB / ', line)
                if result is not None:
                    amfTimeX.append(x * LOOP_INTERVAL)
                    amfCpuY.append(float(result.group('cpu_usage')))
                    amfMemY.append(float(result.group('memory_usage')))
            if line.count('oai-nrf') > 0:
                result = re.search(' (?P<cpu_usage>[0-9\.]+)% *(?P<memory_usage>[0-9\.]+)MiB / ', line)
                if result is not None:
                    nrfTimeX.append(x * LOOP_INTERVAL)
                    nrfCpuY.append(float(result.group('cpu_usage')))
                    nrfMemY.append(float(result.group('memory_usage')))
            if line.count('oai-ausf') > 0:
                result = re.search(' (?P<cpu_usage>[0-9\.]+)% *(?P<memory_usage>[0-9\.]+)MiB / ', line)
                if result is not None:
                    ausfTimeX.append(x * LOOP_INTERVAL)
                    ausfCpuY.append(float(result.group('cpu_usage')))
                    ausfMemY.append(float(result.group('memory_usage')))
            if line.count('oai-udm') > 0:
                result = re.search(' (?P<cpu_usage>[0-9\.]+)% *(?P<memory_usage>[0-9\.]+)MiB / ', line)
                if result is not None:
                    udmTimeX.append(x * LOOP_INTERVAL)
                    udmCpuY.append(float(result.group('cpu_usage')))
                    udmMemY.append(float(result.group('memory_usage')))
            if line.count('oai-udr') > 0:
                result = re.search(' (?P<cpu_usage>[0-9\.]+)% *(?P<memory_usage>[0-9\.]+)MiB / ', line)
                if result is not None:
                    udrTimeX.append(x * LOOP_INTERVAL)
                    udrCpuY.append(float(result.group('cpu_usage')))
                    udrMemY.append(float(result.group('memory_usage')))
            if line.count('oai-smf') > 0:
                result = re.search(' (?P<cpu_usage>[0-9\.]+)% *(?P<memory_usage>[0-9\.]+)MiB / ', line)
                if result is not None:
                    smfTimeX.append(x * LOOP_INTERVAL)
                    smfCpuY.append(float(result.group('cpu_usage')))
                    smfMemY.append(float(result.group('memory_usage')))
            if line.count('oai-spgwu') > 0:
                result = re.search(' (?P<cpu_usage>[0-9\.]+)% *(?P<memory_usage>[0-9\.]+)MiB / ', line)
                if result is not None:
                    spgwuTimeX.append(x * LOOP_INTERVAL)
                    spgwuCpuY.append(float(result.group('cpu_usage')))
                    spgwuMemY.append(float(result.group('memory_usage')))
        # Checking the status of each gnbsim container
        ret = []
        for idx in range(NB_GNBSIM_INSTANCES):
            cmd = f'docker logs omec-gnbsim-{idx + 1} 2>&1 | grep --colour=never "Profile " | grep -v "Waiting for UEs to finish processing" || true'
            tmpRet = run_cmd(cmd, notSilentForFirstTime)
            if tmpRet is None:
                exit(f'\033[0;31m Incorrect/Unsupported executing command "{cmd}"')
            ret.append(str(tmpRet))
        notSilentForFirstTime = True
        allFinished = True
        allPassing = True
        for idx in range(NB_GNBSIM_INSTANCES):
            cnt = ret[idx].count('Profile Status:')
            passing = ret[idx].count('Profile Status: PASS')
            if cnt != NB_PROFILES[idx]:
                allFinished = False
            if passing != NB_PROFILES[idx]:
                allPassing = False
        if allFinished:
            logging.info('\033[0;32m All profiles finished\033[0m....')
            if allPassing:
                logging.info('\033[0;32m All profiles passed\033[0m....')
                status = 0
            else:
                logging.error('\033[0;32m Some profiles failed\033[0m....')
                status = -1
            for idx in range(NB_GNBSIM_INSTANCES):
                print(ret[idx])
            break
        time.sleep(LOOP_INTERVAL)
    cmd = 'docker ps -a'
    res = run_cmd(cmd, False)
    print (res)
    logging.info('Generating a plot for memory usage')
    # Generating a plot for memory usage
    plt.plot(amfTimeX, amfMemY, label='AMF')
    plt.plot(nrfTimeX, nrfMemY, label='NRF')
    plt.plot(ausfTimeX, ausfMemY, label='AUSF')
    plt.plot(udmTimeX, udmMemY, label='UDM')
    plt.plot(udrTimeX, udrMemY, label='UDR')
    plt.plot(smfTimeX, smfMemY, label='SMF')
    if len(spgwuTimeX) > 0:
        plt.plot(spgwuTimeX, spgwuMemY, label='SPGWU')
    plt.legend()
    plt.title('Memory Usage per NF')
    plt.ylabel('MiB')
    plt.xlabel('seconds')
    plt.savefig('oai-cn5g-memory.png')
    plt.cla()
    plt.clf()
    logging.info('Generating a plot for CPU usage')
    # Generating a plot for cpu usage
    plt.plot(amfTimeX, amfCpuY, label='AMF')
    plt.plot(nrfTimeX, nrfCpuY, label='NRF')
    plt.plot(ausfTimeX, ausfCpuY, label='AUSF')
    plt.plot(udmTimeX, udmCpuY, label='UDM')
    plt.plot(udrTimeX, udrCpuY, label='UDR')
    plt.plot(smfTimeX, smfCpuY, label='SMF')
    plt.plot(spgwuTimeX, spgwuCpuY, label='SPGWU')
    plt.legend()
    plt.title('CPU Usage per NF')
    plt.ylabel('%age')
    plt.xlabel('seconds')
    plt.savefig('oai-cn5g-cpu.png')
    if not allFinished:
        logging.error('\033[0;31m Some profiles could not finish\033[0m....')
        for idx in range(NB_GNBSIM_INSTANCES):
            print(ret[idx])
        sys.exit(-1)
    sys.exit(status)

def _parse_args() -> argparse.Namespace:
    """Parse the command line args

    Returns:
        argparse.Namespace: the created parser
    """
    parser = argparse.ArgumentParser(description='Script to check if OMEC-gnbsim deployment went OK.')

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
