# Chronos-WAL
**An Industrial-Grade, Kubernetes-Native Sidecar for PostgreSQL PITR and Observability.**

[![Kubernetes](https://img.shields.io/badge/Kubernetes-Sidecar-blue.svg)](https://kubernetes.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-blue.svg)](https://www.postgresql.org/)
[![AWS S3](https://img.shields.io/badge/AWS-S3-orange.svg)](https://aws.amazon.com/s3/)

**Chronos-WAL** is a specialized infrastructure tool that automates database transaction log (WAL) archiving, compression, and point-in-time recovery (PITR) within a Kubernetes environment. By utilizing the **Sidecar Design Pattern**, it abstracts complex maintenance tasks away from the core database engine.

---

## 🏗 System Architecture
The project moves beyond simple scripting into a professional **Cloud-Native Sidecar Architecture**.



### Key Structural Components:
1.  **Main Container (PostgreSQL):** The primary data engine. It writes WAL logs to a shared mount point.
2.  **Sidecar Container (Chronos-WAL):** A lightweight Python-based daemon that monitors, compresses, and ships those logs to AWS S3.
3.  **Shared Volume (`emptyDir` or `PVC`):** The bridge that allows both containers to access the transaction logs simultaneously without network overhead.

---

## 🚀 Industrial Use Cases

### 1. Zero-Loss Disaster Recovery (DR)
* **The Problem:** Standard daily snapshots leave a "Data Gap." If the DB fails at 11 PM, all data since the midnight backup is lost.
* **The Solution:** Chronos-WAL streams change logs in real-time. In a failure, you restore the last snapshot and "replay" the WAL stream from S3 to recover up to the final millisecond.

### 2. Multi-Region Persistence
* **The Problem:** An entire AWS availability zone or region goes offline.
* **The Solution:** Chronos-WAL ships logs to a cross-region S3 bucket. A new K8s cluster can be spun up in a healthy region and reconstructed using the remote log stream.

### 3. Advanced Observability (LGTM Stack)
* **The Problem:** Traditional database backups are "black boxes." You don't know if they are failing until you need them.
* **The Solution:** Chronos-WAL exposes a `/metrics` endpoint. 
    * **Prometheus** scrapes metrics like `wal_upload_latency`.
    * **Grafana** alerts you if the "Backup Lag" exceeds 60 seconds.

---

## 🛠 Features & Capabilities
| Feature | Implementation | Benefit |
| :--- | :--- | :--- |
| **Atomic Archiving** | `.part` file logic for S3 uploads | Prevents corrupted restores from partial uploads. |
| **Zstd Compression** | Python `zstandard` library | Reduces S3 storage costs by up to 70% with low CPU hit. |
| **Timeline Branching** | `.history` file tracking | Safely handles database restores that create new timelines. |
| **Self-Healing** | K8s Liveness/Readiness probes | K8s restarts the backup daemon automatically if it crashes. |

---

## Project Roadmap (Definition of Done)
To ensure project completion and mastery, the following milestones must be reached:

- [ ] **Phase 1:** Core Python CLI for WAL compression and S3 multi-part uploading.
- [ ] **Phase 2:** Dockerization and `StatefulSet` manifest for the Sidecar Pattern.
- [ ] **Phase 3:** Integration of a Prometheus exporter for real-time backup monitoring.
- [ ] **Phase 4:** A "One-Step" Restore script that parses S3 metadata to calculate recovery chains.

---

## 🔧 Installation (Development)
```bash
# 1. Clone the repository
git clone [https://github.com/your-username/Chronos-WAL.git](https://github.com/your-username/Chronos-WAL.git)

# 2. Build the Sidecar Image
docker build -t chronos-wal:v2.0 .

# 3. Deploy to Kubernetes
kubectl apply -f k8s/statefulset.yaml
