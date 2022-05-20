print("serve 10002")
print("peer 192.168.0.103:10000")
import time
time.sleep(20)
print("name Adder")
print("transfer Alice 102")
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
    print(f"transfer {a} {b} 1")
