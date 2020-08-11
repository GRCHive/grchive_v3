import requests
import os

def pullData(url):
    token = os.environ["GITLAB_API_TOKEN"]

    r = requests.get(url=url, headers={
        "PRIVATE-TOKEN" : token  
    })
    data = r.json()
    return data, r.links["next"]["url"] if "next" in r.links else None

initialUrl = "https://gitlab.com/api/v4/projects/grchive%2Fgrchive-v3/variables"
data, nextUrl = pullData(initialUrl)

while True:
    for datum in data:
        key = datum['key']
        val = datum['value'].replace("\"", "\\\"")
        print("export {}=\"{}\"".format(key, val))

    if nextUrl == None:
        break

    data, nextUrl = pullData(nextUrl)
