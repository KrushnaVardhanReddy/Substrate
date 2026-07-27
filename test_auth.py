import requests
import subprocess

out = subprocess.check_output('podman exec test_forgejo5 su git -c "gitea admin user generate-access-token --username adminuser --token-name pytoken --raw"', shell=True)
TOKEN = out.decode('utf-8').strip().split('\n')[-1].strip()
print("Using token:", TOKEN)
URL = "http://127.0.0.1:3000/api/v1/user"

r2 = requests.get(URL, headers={"Authorization": f"token {TOKEN}"})
print("token:", r2.status_code, r2.text)

r1 = requests.get(URL, auth=("adminuser", "Admin123!"))
print("Basic:", r1.status_code, r1.text)

