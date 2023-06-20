import queue
import sys

import requests
import time
from concurrent.futures import ThreadPoolExecutor


def get_ips():
    headers = {
        'authority': 'unpkg.com',
        'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36',
    }

    response = requests.get('https://unpkg.com/@hcfy/google-translate-ip@1.0.0/ips.txt', headers=headers)
    ips = response.text.split("\n")
    return ips


def check_ip(ip:str, q: queue.Queue):
    HOST = 'translate.googleapis.com'
    TESTIP_FORMAT = 'https://{}/translate_a/single?client=gtx&sl=en&tl=fr&q=a'
    url = TESTIP_FORMAT.format(ip)
    headers = {'Host': HOST}

    start_time = time.time()
    try:
        response = requests.get(url, headers=headers, timeout=2.5, verify=False)
    except requests.exceptions.ConnectTimeout as e:
        # print(f"超时：【{ip}】")
        return ""
    use_time = time.time() - start_time
    print('****************')
    print(f"ip:【{ip}】|耗时：【{use_time}】")
    print("****************")
    q.put(ip)
    return ip


def write_hosts(ip: str):
    hosts_path = r'C:\Windows\System32\drivers\etc\hosts' if sys.platform == 'win32' else '/etc/hosts'
    with open(hosts_path, "r", encoding="utf-8") as f:
        lines = f.readlines()
    change_ip = "{} translate.googleapis.com\n"
    for index, line in enumerate(lines):
        if line.find("translate.googleapis.com") != -1:
            change_ip = change_ip.format(ip)
            lines[index] = change_ip
    with open(hosts_path, "w", encoding="utf-8") as f:
        f.writelines(lines)


if __name__ == '__main__':
    ips = get_ips()
    q = queue.Queue()
    pool = ThreadPoolExecutor(max_workers=64)
    for ip in ips:
        pool.submit(check_ip, ip, q)
    pool.shutdown()
    if not q.empty():
        ip = q.get()
        write_hosts(ip)
        print(f"改写：【{ip}】")
