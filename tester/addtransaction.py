print("serve 10002", flush=True)
print("peer 192.168.0.102:10000", flush=True)
import time
time.sleep(10)
print("name Adder", flush=True)
print("transfer Alice 102", flush=True)
people = ["Alice", "Bob", "Charlie", "David"]
warmup = 10
while True:
    import time
    time.sleep(0.5)
    import random
    a = people[random.randint(0, len(people) - 1)]
    b = people[random.randint(0, len(people) - 1)]
    if warmup > 0 or random.random() > 0.3:
        a = "Alice"
    print(f"transfer {a} {b} 1", flush=True)
