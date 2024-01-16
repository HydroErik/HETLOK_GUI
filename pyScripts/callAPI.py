"""Driver to demo API calls in short concise code"""

import requests
#import os


preAmble = "http://127.0.0.1:8080/"
#query = "pipes?pipeType=TRANSFORMER&adminProperties=Yes" #pipe call w/t admin
#query = "pipes?pipeType=TRANSFORMER&adminProperties=No"  #pipe call no admin
query = "selectLists?selectList=0&noItems=true"

table = "credentialType"
operation = "allRecords"
#recordId = 
#name = 
#clientId = 0
#jobId = 
#startTimestamp = 
#endTimestamp =



#query = f"query?table={table}&operation={operation}"


url = preAmble + query
print(url)

r  = requests.get(url=url)

print(f"Status Code: {r.status_code}\nText: {r.text}")
