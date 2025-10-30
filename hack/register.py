import requests

API_URL = "http://localhost/providers"

provider_data = {
    "name": "Ansible Provider",
    "type": "ansible",
    "endpoint": "http://localhost:1234/api",
    "metadata": {
        "region": "eu-central-1",
        "supportedServices": ["virtual_machine", "database"]
    }
}

headers = {
    "Content-Type": "application/json",
}

response = requests.post(API_URL, json=provider_data, headers=headers)

if response.status_code == 201:
    provider = response.json()
    print("Registered provider:", provider["id"], provider["name"])
else:
    print("Failed:", response.status_code, response.text)
