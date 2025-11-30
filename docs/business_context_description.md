# 1. Business Goal (The Vision & Mission)

**Primary Vision:** To establish the world's leading open, decentralized ecosystem for **distributing and managing** indie games on low-cost IoT hardware. We are building the "Steam" for the ESP32 era.

**Our Mission:**

- **Create the Standard:** To develop an open-source platform that transforms common IoT devices (like the M5 series) into connected consoles. We move beyond simple "offline" sketches to a professional-grade **Connected Lifecycle**: creating, publishing, updating, and tracking.
    
- **Lower the Barrier:** To provide a powerful **ESP-IDF Component** that abstracts away the complexity of secure networking. A developer should focus on gameplay, whilst the framework handles FreeRTOS tasks, TLS, and memory management.
    
- **Build the Infrastructure:** To provide a reference implementation for the "Crisp Cloud" (the backend), built on high-performance **Go (Golang)** and **Azure Kubernetes Service**, capable of handling millions of devices.
    
- **Foster a Community:** To create a global social graph where players can compete on leaderboards, unlock achievements, and track their stats via a web portal, bridging the physical device with the digital community.
    

# 2. Market Landscape

We are creating a new category: **The Connected Indie Console**. We solve the distribution and engagement problems that plague the current DIY scene.

|**Market**|**Existing Problem**|**The "Crisp" Solution**|
|---|---|---|
|**Closed Consoles** (Switch)|A "walled garden." High barrier to entry, strict curation, no access to hardware internals.|**Radical Openness.** Anyone can upload a `.bin` file. The platform handles the delivery (FOTA). No permission needed.|
|**Indie Handhelds** (Playdate)|Beautiful but isolated. Experiences are mostly offline. Updating games is often manual or clunky.|**Connected by Default.** "Crisp" devices are always syncing. Updates are pushed over-the-air. Stats are uploaded instantly.|
|**DIY Platforms** (Arduino/Pi)|High friction. To share a game, you often share source code. Users must compile it themselves.|**Seamless Distribution.** "Crisp" acts as an App Store. Players just download compiled binaries. Developers get aggregated telemetry.|

Our Unique Selling Proposition (USP):

"Crisp" is the Infrastructure-as-a-Service for IoT Games. It allows a lone developer to build a game that has the distribution power and social features (leaderboards, cloud saves) of a AAA studio, running on $20 hardware.

# 3. Scope (The Vision for v1.0)

### What "Crisp" Is (In Scope):

1. **"Crisp Firmware" (The Device):**
    
    - A professional-grade **Native C++ Framework** built directly on **ESP-IDF** (Espressif IoT Development Framework).
        
    - Leverages **FreeRTOS** to efficiently manage the **Single-Player Game Loop** (Core 1) and background **Network Tasks** (Core 0).
        
    - Handles secure **MQTT Telemetry** (sending scores/usage) and **HTTPS FOTA** (downloading new games) transparently.
        
2. **"Crisp Cloud" (The Backend):**
    
    - **Monolithic Go Service:** A high-performance backend written in **Go (Golang)** running on **Azure Kubernetes Service (AKS)**.
        
    - **FOTA Engine:** Manages firmware versioning and streams `.bin` files to devices via secure channels.
        
    - **Telemetry Ingest:** Subscribes to **MQTT** topics to process game stats in real-time.
        
3. **"Crisp Portal" (The Web Interface):**
    
    - **For Players:** A social dashboard (HTML/Go Templates) to view profiles, achievement badges, and global leaderboards.
        
    - **For Developers:** An Admin Panel to upload new game binaries (`.bin`) and view fleet analytics.
        

### What "Crisp" Is Not (Out of Scope):

- **Real-Time Multiplayer:** We are not synchronizing player movement tick-by-tick (no UDP game state sync). The focus is on _asynchronous_ competition (High Scores).
    
- **Hardware Manufacturer:** We provide the software stack for existing ESP32 devices (M5Stack, Custom Boards).
    

# 4. Use Cases (Ecosystem Scenarios)

### Scenario 1: The Developer (DevOps Flow)

- **Actor:** An embedded systems student or indie dev.
    
- **Goal:** To publish a "Flappy Bird" clone and update it with a bug fix.
    
- **How it works:**
    
    1. She writes the game in C++ using the **Crisp ESP-IDF Component** and builds it using `idf.py build`.
        
    2. She logs into the **Crisp Admin Portal** (hosted on Azure) and uploads the generated `game.bin`.
        
    3. She notices a bug, fixes it, recompiles, and pushes `v2` to the portal via a CI/CD pipeline (GitHub Actions).
        
    4. The system automatically marks `v2` as the latest stable release.
        

### Scenario 2: The Player (The Consumer)

- **Actor:** A gamer with an M5StickC.
    
- **Goal:** To play a new game and compete with friends.
    
- **How it works:**
    
    1. He turns on his device. The device boots and checks the **FOTA Service** via HTTPS.
        
    2. The device sees a new game version is available and downloads it automatically to the OTA partition.
        
    3. He plays the game offline on the bus.
        
    4. When he gets home (Wi-Fi), the device silently pushes his high score via **MQTT** in the background.
        
    5. He opens the **Crisp Web Portal** on his phone and sees he is ranked #5 globally.
        

### Scenario 3: The Data Analyst

- **Actor:** The platform owner.
    
- **Goal:** To understand player engagement.
    
- **How it works:**
    
    1. They open the **Grafana/Admin Dashboard**.
        
    2. They see real-time metrics processed by the **Go Backend**: "Daily Active Devices", "Average Session Time", "Crash Reports".
        
    3. All data is queried securely from the **PostgreSQL** database running in the Kubernetes cluster.
        

# 5. User Stories

- **As a Creator,** I want to build my game using **ESP-IDF**, so that I have full control over memory and tasks without the overhead of the Arduino framework.
    
- **As a Backend Engineer,** I want a robust **Go (Golang)** application running in **Docker**, so that I can easily deploy it to **Azure Kubernetes Service**.
    
- **As a Player,** I want to view my gaming history and stats on a website, so that I can share my achievements with friends.
    
- **As a Hardware Owner,** I want my device to automatically update its firmware, so that I always have the latest content and bug fixes.
    
- **As a System Admin,** I want to use **Azure Files** and **Postgres**, so that my game data and binary files are persistent and backed up reliably.