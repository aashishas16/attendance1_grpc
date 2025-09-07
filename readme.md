Perfect 👍 thanks for sharing the responses, this makes it clear how your API works.
From what you’ve tested:

* ✅ **Check-in works** at `POST /v1/checkin`
* ❌ `POST /v1/checkout` is **not valid** (404)
* ✅ Checkout works with **PUT /v1/checkout/{id}\`**
* ✅ Fetch a single record works with **GET /v1/attendance/{id}\`**
* ✅ Fetch all records works with **GET /v1/attendance\`**

---

### ✅ Correct `curl` commands (all 4)

**1. Check-in**

```bash
curl -X POST http://localhost:8080/v1/checkin \
  -H "Content-Type: application/json" \
  -d '{"user_id":"1", "username":"aashish1"}'
```

**2. Checkout (by record ID)**

```bash
curl -X PUT http://localhost:8080/v1/checkout/68bdad39bc09a5e2a95c53b5 \
  -H "Content-Type: application/json" \
  -d '{}'
```

**3. Get single attendance record (by record ID)**

```bash
curl -X GET http://localhost:8080/v1/attendance/68bdad39bc09a5e2a95c53b5 \
  -H "Content-Type: application/json"
```

**4. Get all attendance records**

```bash
curl -X GET http://localhost:8080/v1/attendance \
  -H "Content-Type: application/json"
```

---

👉 Now you have the **working set of 4 core endpoints**:

* **Check-in**
* **Checkout**
* **Get one record**
* **Get all records**

Do you also want me to make a **ready-to-run bash script** (`attendance_test.sh`) that runs all these 4 commands in sequence and prints the results?
# attendance1_grpc
