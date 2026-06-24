# 📱 Mobile Networks – Module 1: Study Guide
**Course:** RXS1404 | Burkina Institute of Technology  
**Instructor:** B. Tanguy KABORE

---

## Q1. What is a Mobile Network? Define its 4 key characteristics.

**A mobile network** is a wireless telecommunications infrastructure that enables voice and data communication using the radio frequency spectrum while providing mobility to users.

Its 4 key characteristics are:

| Feature | Description |
|---|---|
| **Mobility** | Seamless handover between cells as the user moves |
| **Coverage** | Wide geographical area served by multiple base stations |
| **Capacity** | Ability to handle multiple simultaneous users |
| **Quality of Service (QoS)** | Reliable connections maintained throughout communication |

---

## Q2. What are the 4 main components of a mobile network? Briefly define each.

| Component | Role |
|---|---|
| **Mobile Station (MS)** | User's terminal device (phone, tablet, IoT) — the **edge** of the network |
| **Base Station (BS)** | Radio tower providing the wireless link between MS and the network |
| **Core Network (CN)** | The "brain" — handles control, routing, authentication, and billing |
| **Support Systems (OSS/BSS)** | Management tools for network operations and business functions |

> **Key interface rule:** MS ↔ BS = **Radio (air)** | BS ↔ Core = **Backhaul (wired)** | Core ↔ Internet = **IP transport**

---

## Q3. Describe the evolution of mobile networks from 1G to 5G (timeline + key technology per generation).

| Generation | Year | Technology | Key Services |
|---|---|---|---|
| **1G** | 1980s | Analog FDMA | Voice only (AMPS, NMT, TACS) |
| **2G** | 1991 | Digital GSM (TDMA) | Voice + SMS + MMS + GPRS/EDGE |
| **3G** | 2001 | WCDMA / UMTS | Mobile internet, video calls (up to 42 Mbps HSPA+) |
| **4G** | 2009 | OFDMA / LTE | High-speed all-IP (up to 1 Gbps, <10ms latency) |
| **5G** | 2020 | NR (New Radio) | Ultra-fast, low latency, IoT, network slicing |
| **6G** | 2030s | TBD | Still in research phase |

---

## Q4. What is a Mobile Station (MS)? List its primary functions and key hardware/software components.

An MS is a **mobile terminal used by subscribers to access the network** (smartphones, tablets, laptops with 4G/5G cards, IoT sensors).

**Primary roles:**
- Transmit and receive data
- Move across cells (mobility / handover)
- Authenticate on the network
- Measure signal quality

**Hardware:** Antenna, Radio Transceiver (RF module), Baseband Processor, Application Processor, Battery  
**Software:** OS (Android/iOS), Radio Protocol Stack, SIM Application Toolkit, User Applications

> The MS contains a **SIM/eSIM** (Subscriber Identity Module) and supports multiple frequency bands.

---

## Q5. What is a Base Station (BS)? How does its naming change across generations?

A BS provides the **wireless link between mobile devices and the network**. It covers a geographic area (cell), manages radio access, and handles handover.

| Technology | Base Station Name | Architecture |
|---|---|---|
| GSM (2G) | BTS (Base Transceiver Station) | Circuit-switched |
| UMTS (3G) | Node B | Packet + circuit |
| LTE (4G) | eNodeB (evolved Node B) | All-IP, flat |
| 5G NR | gNodeB (next-gen Node B) | Virtualized, low latency |
| Wi-Fi | Access Point (AP) | Local area only |

A BS is composed of: **Antenna System** (MIMO arrays), **Radio Unit (RU)** (signal amplification), and **Baseband Unit (BBU)** (digital processing). The BBU and RU are connected via a **fronthaul** link; the BS connects to the Core via a **backhaul** link.

---

## Q6. What is a Cell? What is its theoretical shape and what are the 4 cell types by size?

A **cell** is the geographic area covered by one base station. It enables **frequency reuse** — the fundamental principle that allows the same radio frequencies to be reused across a network.

- **Theoretical shape:** Hexagonal (for perfect tiling with no gaps/overlaps)
- **Reality:** Irregular, shaped by terrain and obstacles

| Cell Type | Radius | Typical Use |
|---|---|---|
| **Macro cell** | 1–30 km | Rural areas, highways |
| **Micro cell** | 100 m – 1 km | Urban areas |
| **Pico cell** | 10–100 m | Indoor hotspots |
| **Femto cell** | 10–50 m | Home / office |

---

## Q7. Explain the concept of Frequency Reuse and the Cluster / Reuse Factor K.

**Problem:** Radio spectrum is a limited resource. If neighboring cells use the same frequency, they cause **co-channel interference**.

**Solution:** Reuse the same frequencies only at a **sufficient distance** to minimize interference.

**Key definitions:**
- **Cluster:** A group of cells that collectively use all available frequencies exactly once
- **K (reuse factor):** The number of cells in one cluster
- If there are **N total frequencies**, each cell receives **N/K** frequencies

**Trade-off:**
- **Small K** → more frequencies per cell → **high capacity** but **high interference**
- **Large K** → fewer frequencies per cell → **low interference** but **reduced capacity**

**Typical values:**
- GSM (2G): K = 4 or 7
- Dense urban: K = 3 or 4
- Rural: K = 7 or 12

---

## Q8. What is the formula for K and how do you calculate it? Give 3 examples.

For hexagonal cells to tile perfectly, K must satisfy:

$$K = i^2 + ij + j^2$$

where **i** and **j** are non-negative integers (≥ 0).

| i | j | Calculation | K |
|---|---|---|---|
| 1 | 0 | 1 + 0 + 0 | **1** |
| 1 | 1 | 1 + 1 + 1 | **3** |
| 2 | 0 | 4 + 0 + 0 | **4** |
| 2 | 1 | 4 + 2 + 1 | **7** |
| 2 | 2 | 4 + 4 + 4 | **12** |

> The formula guarantees a **regular hexagonal geometry** where each cluster can be repeated infinitely.

---

## Q9. What is the Reuse Distance D? Give its formula and calculate D/R for K = 3, 4, and 7.

**Reuse Distance D** is the minimum distance between two co-channel cells (cells using the same frequency) to avoid interference.

$$D = R\sqrt{3K} \quad \Rightarrow \quad \frac{D}{R} = \sqrt{3K}$$

Where:
- **D** = distance between co-channel cell centers
- **R** = cell radius
- **K** = reuse factor

| K | D/R Calculation | Result |
|---|---|---|
| K = 3 | √(3×3) = √9 | **3** |
| K = 4 | √(3×4) = √12 | **≈ 3.46** |
| K = 7 | √(3×7) = √21 | **≈ 4.58** |
| K = 12 | √(3×12) = √36 | **6** |

> A larger D/R means co-channel cells are farther apart → less interference.

---

## Q10. What is Sectorization? What are its benefits?

**Sectorization** is the technique of dividing a single cell into **sectors** using **directional antennas** rather than one omnidirectional antenna.

- Typical configuration: **3 sectors of 120°** each
- Each sector behaves like a separate "mini-cell" with its own frequency allocation

**Benefits:**
- Increased capacity per cell site
- Better frequency reuse efficiency
- Improved signal quality
- Reduced co-channel interference

---

## Q11. What are the functions of the Core Network? Distinguish Control Plane vs User Plane.

The Core Network is the **central, wired, IP-based part** of the mobile network. It handles no radio communication — purely IP/fiber-based transport.

| Control Plane | User Plane |
|---|---|
| Authentication of subscribers | Data routing |
| Mobility management | QoS management |
| Session management | Internet connectivity |
| Policy enforcement | Voice call handling |
| Security (encryption keys) | Billing data collection |

**Core evolution by generation:**

| Generation | Core | Architecture |
|---|---|---|
| 2G (GSM) | NSS: MSC, VLR, HLR, AuC | Circuit-switched |
| 2.5G (GPRS) | SGSN, GGSN | Packet overlay |
| 3G (UMTS) | MSC + SGSN/GGSN | Combined CS/PS |
| 4G (LTE) | EPC: MME, SGW, PGW, HSS | All-IP, flat |
| 5G | 5GC: AMF, SMF, UPF, UDM | Service-based, cloud-native |

---

## Q12. What are OSS and BSS? How do they differ?

Both are **Support Systems** — they carry **management traffic only**, never user data.

| System | Full Name | Responsibilities |
|---|---|---|
| **OSS** | Operations Support System | Network management, service assurance, fault correlation, network inventory, performance analysis |
| **BSS** | Business Support System | Customer care, billing systems, order management, product catalog, partner management, fraud detection |

> Together they form the **management plane** of the mobile network.

---

## Q13. Walk through the complete data flow when a mobile user makes a phone call.

This example illustrates how all 4 network components work together:

1. **MS** sends a call request via radio to **BS**
2. **BS** forwards the signaling to the **Core Network**
3. **Core** authenticates the user (checks HSS/UDM database)
4. **Core** establishes the session and allocates resources
5. **Core** routes the call to the destination
6. **Support Systems (OSS/BSS)** log the event for billing
7. Voice data then flows: **MS ↔ BS ↔ Core ↔ Destination**

> This shows the separation of roles: MS is the **endpoint**, BS handles **radio access**, Core handles **intelligence**, and Support Systems handle **management and billing**.

---

*Bon courage pour ton examen ! 🎯*
