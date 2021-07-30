import threading 
import requests 



def req():
    requests.get("http://localhost:4000/")

for i in range(10000):
    t = threading.Thread(target=req)
    t.start()