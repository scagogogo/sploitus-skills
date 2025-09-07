

请求：
```curl
curl 'https://sploitus.com/search' \
  -H 'accept: application/json' \
  -H 'accept-language: zh-CN,zh;q=0.9' \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/json' \
  -b '_ym_uid=1743360463972135790; _ym_d=1743360463; _ym_isad=1; cf_clearance=6IvGezkOrnZYwNQ0mbtujJPgig8jQ7cyIybccK3vDjE-1743360468-1.2.1.1-vYkSgIymV9AGXfPsy8GdxXFzRL_RwaXoiAfhY827lV4wOkQhOwa.u1mo2FJ4uo3i0xE2qJh5VSLZRxgX_TTCbMTE36gJnQGRbocmT14ZLvBSIzxRsU25SV9burCJp_EQr2VkY93xemXQL6mUGVX63DNYOscz3874_fpHy7JW4WJzQeUnUtfYcytw0tjKyPVB_aemeiPS6K3MSRg4HC1iNyxNYdFF_1Bru_X31VcY7NsL4Ns3bLSu0xZn11OWpargRP4JCGtDWGlTInZEY4IXShKZI0v2EVjbPza8pIJzDpBWqAS41SKB_asWLS.Wd9k2Z8pW.wYeColtC8FqWBKqWeyYRHpABoRjFfZBHQYwP0o; _ym_hostIndex=0-9%2C1-0' \
  -H 'origin: https://sploitus.com' \
  -H 'pragma: no-cache' \
  -H 'priority: u=1, i' \
  -H 'referer: https://sploitus.com/?query=%20CVE-2025-1316' \
  -H 'sec-ch-ua: "Chromium";v="134", "Not:A-Brand";v="24", "Google Chrome";v="134"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "macOS"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-origin' \
  -H 'user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36' \
  --data-raw '{"type":"exploits","sort":"default","query":" CVE-2025-1316","title":false,"offset":0}'
```

翻页请求：
```curl
curl 'https://sploitus.com/search' \
  -H 'accept: application/json' \
  -H 'accept-language: zh-CN,zh;q=0.9' \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/json' \
  -b '_ym_uid=1743360463972135790; _ym_d=1743360463; _ym_isad=1; cf_clearance=6IvGezkOrnZYwNQ0mbtujJPgig8jQ7cyIybccK3vDjE-1743360468-1.2.1.1-vYkSgIymV9AGXfPsy8GdxXFzRL_RwaXoiAfhY827lV4wOkQhOwa.u1mo2FJ4uo3i0xE2qJh5VSLZRxgX_TTCbMTE36gJnQGRbocmT14ZLvBSIzxRsU25SV9burCJp_EQr2VkY93xemXQL6mUGVX63DNYOscz3874_fpHy7JW4WJzQeUnUtfYcytw0tjKyPVB_aemeiPS6K3MSRg4HC1iNyxNYdFF_1Bru_X31VcY7NsL4Ns3bLSu0xZn11OWpargRP4JCGtDWGlTInZEY4IXShKZI0v2EVjbPza8pIJzDpBWqAS41SKB_asWLS.Wd9k2Z8pW.wYeColtC8FqWBKqWeyYRHpABoRjFfZBHQYwP0o; _ym_hostIndex=0-30%2C1-0' \
  -H 'origin: https://sploitus.com' \
  -H 'pragma: no-cache' \
  -H 'priority: u=1, i' \
  -H 'referer: https://sploitus.com/?query=%20CVE-2023-1234' \
  -H 'sec-ch-ua: "Chromium";v="134", "Not:A-Brand";v="24", "Google Chrome";v="134"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "macOS"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-origin' \
  -H 'user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36' \
  --data-raw '{"type":"exploits","sort":"default","query":" CVE-2023-1234","title":false,"offset":20}'
```

响应：
```json
{
    "exploits": [
        {
            "title": "Exploit for CVE-2025-1316",
            "score": 9.3,
            "href": "https://github.com/slockit/CVE-2025-1316",
            "type": "githubexploit",
            "published": "2025-03-29",
            "id": "0147E6AA-6963-51CE-90F9-420346FA917B",
            "source": "## https://sploitus.com/exploit?id=0147E6AA-6963-51CE-90F9-420346FA917B\n# CVE-2025-1316\n\n> Run as root\n\nEdimax IC-7100 does not properly neutralize requests. An attacker can create specially crafted requests to achieve remote code execution on the device\n\n\n# Install\n\n```\nsudo apt update\nsudo apt install git\ngit clone https://github.com/slockit/CVE-2025-1316.git\ncd CVE-2025-1316\nchmod +x CVE-2025-1316\nchmod +x install.sh\nsudo bash install.sh\n```\n\n# Usage\n```\n./CVE-2025-1316 [https://.com/]\n```\n```\n./CVE-2025-1316 -l hosts.txt -t 30\n```",
            "language": "MARKDOWN"
        },
        {
            "title": "Edimax IP Camera NTP_serverName command injection",
            "score": 9.3,
            "href": "https://download.saintcorporation.com/cgi-bin/exploit_info/edimax_ip_camera_ntp_servername",
            "type": "saint",
            "published": "2025-03-21",
            "id": "SAINT:2CEDD0194C77120545A6315E534CFE66",
            "source": "## https://sploitus.com/exploit?id=SAINT:2CEDD0194C77120545A6315E534CFE66\nAdded: 03/21/2025  \nCVE: CVE-2025-1316  \n\n\n### Background\n\nEdimax IP Cameras are a product line of security cameras which send video footage over an IP network. \n\n### Problem\n\nA command injection vulnerability in the `**NTP_serverName**` POST parameter of an update request allows remote attackers to execute arbitrary commands. This vulnerability can be exploited using a well known default password. \n\n### Resolution\n\nMinimize network exposure of the device, and ensure that it is not reachable from the Internet. Use a VPN if remote access is needed. \n\n### References\n\nhttps://www.cisa.gov/news-events/ics-advisories/icsa-25-063-08   \n\n\n### Limitations\n\nExploit only works if the default device password is unchanged.",
            "language": "MARKDOWN"
        }
    ],
    "exploits_total": 2
}
```


