---

#triple-s — Simple Storage Service (S3 Clone in Go)

**triple-s** is a simplified, fully functional implementation of an object storage service inspired by **Amazon S3**, written entirely in **Go** using only standard packages.

It provides a **RESTful HTTP API** that allows clients to:

* Create, list, and delete **buckets**
* Upload, retrieve, and delete **objects**
* Manage and persist **metadata** for both buckets and objects in **CSV files**
* Interact using **XML responses**, conforming to S3 response conventions

This project demonstrates core concepts of **HTTP servers**, **REST API design**, **networking**, and **persistent storage**, serving as a foundational exploration of how modern cloud storage services are built.

---

## 🚀 Features

### 🪣 Bucket Management

* **Create a Bucket** (`PUT /{BucketName}`)
  Validates and creates a new bucket with proper metadata.
* **List Buckets** (`GET /`)
  Returns an XML list of all existing buckets and their metadata.
* **Delete a Bucket** (`DELETE /{BucketName}`)
  Deletes an existing, empty bucket from storage.

### 📦 Object Operations

* **Upload Object** (`PUT /{BucketName}/{ObjectKey}`)
  Uploads or overwrites an object and stores its metadata.
* **Retrieve Object** (`GET /{BucketName}/{ObjectKey}`)
  Streams object content with appropriate headers.
* **Delete Object** (`DELETE /{BucketName}/{ObjectKey}`)
  Deletes an object and updates metadata.

### ⚙️ Additional Capabilities

* XML-based responses to ensure S3-like communication.
* Metadata persistence using **CSV files** for simplicity and portability.
* Input validation for bucket names and object keys.
* Graceful error handling with correct HTTP status codes.
* Lightweight and dependency-free — uses only Go’s standard library.

---

## 🧠 Architecture Overview

The triple-s service is composed of three main layers:

### 1. **HTTP Server Layer**

Handles incoming HTTP requests using the `net/http` package.
Parses routes, manages URL patterns, and dispatches requests to the appropriate handlers for buckets or objects.

### 2. **Storage Layer**

Manages file system operations:

* Buckets are represented as directories.
* Objects are stored as files within those directories.
* Metadata (for both buckets and objects) is stored in `.csv` files.

### 3. **Metadata Layer**

Each metadata CSV file maintains structured information about the resources:

#### `buckets.csv`

| Name   | CreationTime         | LastModifiedTime     | Status |
| ------ | -------------------- | -------------------- | ------ |
| photos | 2025-10-15T12:31:45Z | 2025-10-15T12:31:45Z | active |

#### `data/{bucket}/objects.csv`

| ObjectKey  | Size   | ContentType | LastModified         |
| ---------- | ------ | ----------- | -------------------- |
| sunset.png | 214532 | image/png   | 2025-10-15T13:02:15Z |

---

## 🧱 Directory Structure

```bash
project-root/
├── triple-s             # Compiled binary
├── main.go              # Entry point
├── internal/
│   ├── server.go        # HTTP server setup
│   ├── bucket.go        # Bucket management logic
│   ├── object.go        # Object operations
│   └── xmlresponse.go   # XML response helpers
├── data/                # Base directory for storage
│   ├── buckets.csv      # Bucket metadata
│   └── {bucket-name}/
│       ├── objects.csv  # Object metadata for this bucket
│       └── {object-key} # Actual object files
└── README.md
```

---

## ⚡ Usage

### Build

```bash
$ go build -o triple-s .
```

### Run

```bash
$ ./triple-s -port 8080 -dir ./data
```

or display help:

```bash
$ ./triple-s --help
```

#### Output:

```
Simple Storage Service.

Usage:
  triple-s [-port <N>] [-dir <S>]
  triple-s --help

Options:
  --help     Show this screen.
  --port N   Port number
  --dir S    Path to the storage directory
```

---

## 🌐 API Endpoints

### 🪣 Bucket Management

#### Create Bucket

**PUT /{BucketName}**

```bash
curl -X PUT http://localhost:8080/my-bucket
```

**Responses:**

* `200 OK` — Bucket created
* `400 Bad Request` — Invalid name
* `409 Conflict` — Bucket already exists

---

#### List Buckets

**GET /**

```bash
curl http://localhost:8080/
```

**Response (XML Example):**

```xml
<ListAllMyBucketsResult>
  <Buckets>
    <Bucket>
      <Name>my-bucket</Name>
      <CreationDate>2025-10-28T14:22:00Z</CreationDate>
    </Bucket>
  </Buckets>
</ListAllMyBucketsResult>
```

---

#### Delete Bucket

**DELETE /{BucketName}**

```bash
curl -X DELETE http://localhost:8080/my-bucket
```

**Responses:**

* `204 No Content` — Bucket deleted
* `404 Not Found` — Bucket does not exist
* `409 Conflict` — Bucket not empty

---

### 📦 Object Operations

#### Upload Object

**PUT /{BucketName}/{ObjectKey}**

```bash
curl -X PUT -T ./sunset.png \
  -H "Content-Type: image/png" \
  http://localhost:8080/photos/sunset.png
```

**Responses:**

* `200 OK` — Object uploaded
* `404 Not Found` — Bucket not found
* `400 Bad Request` — Invalid key

---

#### Retrieve Object

**GET /{BucketName}/{ObjectKey}**

```bash
curl -O http://localhost:8080/photos/sunset.png
```

**Responses:**

* `200 OK` — Returns binary data
* `404 Not Found` — Bucket or object missing

---

#### Delete Object

**DELETE /{BucketName}/{ObjectKey}**

```bash
curl -X DELETE http://localhost:8080/photos/sunset.png
```

**Responses:**

* `204 No Content` — Object deleted
* `404 Not Found` — Object or bucket missing

---

## 🧩 Validation Rules

### Bucket Names

* 3–63 characters
* Only lowercase letters, numbers, hyphens, and dots
* Cannot resemble an IP (e.g. `192.168.0.1`)
* Cannot start or end with `-` or `.`
* No consecutive `..` or `--`

### Object Keys

* Must be non-empty
* Cannot contain path traversal (`../`)
* Must fit within file system name limits

---

## 🔐 Error Handling

The server handles all errors gracefully and never crashes under invalid requests.

| Error Condition     | HTTP Status | Example                                           |
| ------------------- | ----------- | ------------------------------------------------- |
| Invalid bucket name | 400         | `<Error><Code>InvalidBucketName</Code></Error>`   |
| Bucket exists       | 409         | `<Error><Code>BucketAlreadyExists</Code></Error>` |
| Bucket not found    | 404         | `<Error><Code>NoSuchBucket</Code></Error>`        |
| Object not found    | 404         | `<Error><Code>NoSuchKey</Code></Error>`           |
| Bucket not empty    | 409         | `<Error><Code>BucketNotEmpty</Code></Error>`      |
| Internal I/O error  | 500         | `<Error><Code>InternalError</Code></Error>`       |

All error responses are XML-formatted per S3 conventions.

---

## 🧠 Design Decisions

1. **Go Standard Library Only:**
   The entire implementation relies solely on the `net/http`, `encoding/xml`, `encoding/csv`, and `os` packages — no external dependencies.

2. **CSV for Metadata:**
   CSV files provide a simple, readable persistence layer ideal for lightweight prototypes without a database.

3. **Modular Structure:**
   Clear separation of concerns between HTTP handlers, file management, and XML response generation.

4. **Graceful Error Handling:**
   No panics — all recoverable conditions return proper HTTP responses.

5. **Scalable Foundation:**
   The design can be extended to include authentication, versioning, and multipart uploads in the future.

---

## 🧪 Example Workflow

```bash
# Start server
$ ./triple-s -port 8080 -dir ./data

# Create bucket
$ curl -X PUT http://localhost:8080/photos

# Upload image
$ curl -X PUT -T ./image.png -H "Content-Type: image/png" http://localhost:8080/photos/image.png

# Retrieve image
$ curl -O http://localhost:8080/photos/image.png

# Delete image
$ curl -X DELETE http://localhost:8080/photos/image.png

# Delete bucket
$ curl -X DELETE http://localhost:8080/photos
```

---

## 🧾 Lessons Learned

Building **triple-s** was a deep dive into:

* The **internals of HTTP servers** and **routing** in Go
* Understanding **RESTful principles** and **stateless design**
* Implementing **persistent metadata management**
* Designing **robust error handling** for distributed systems
* Building reliable **cloud storage primitives**

This project solidified my understanding of how services like **Amazon S3** operate at their core, translating complex cloud concepts into practical, working code.

---

## 🧰 Technologies Used

* **Language:** Go (1.22+)
* **Core Packages:** `net/http`, `os`, `io`, `encoding/xml`, `encoding/csv`, `regexp`, `time`, `fmt`
* **Data Format:** XML for responses, CSV for persistence
* **Testing:** Manual API testing via `curl` and Postman

---

## 📄 License

This project was built for educational purposes.
All design and implementation decisions are my own.
Inspired by the architecture of Amazon S3.

---
