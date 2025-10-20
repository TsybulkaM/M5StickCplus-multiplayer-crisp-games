### 1. Business Goal (The Vision & Mission)

**Primary Vision:** To establish the world's leading open, decentralized ecosystem for creating and sharing "connected experiences" on low-cost, connected IoT hardware.

**Our Mission:**

1. **Create the Standard:** To develop and open-source — a lightweight, high-performance platform that transforms common IoT devices (like the M5 series) into powerful, _natively-networked_ interactive consoles.
    
2. **Lower the Barrier:** To provide a powerful, yet simple, SDK that abstracts away the complexity of real-time networking. A developer should be able to create a secure, real-time multiplayer experience.
    
3. **Build the Infrastructure:** To provide a reference implementation for the "Crisp Hub" (the backend), built on accessible, ubiquitous technologies (**Node.js/TypeScript** and/or **Python/Django**) that any developer can deploy, modify, and scale.
    
4. **Foster a Community:** To build a global community of "makers," developers, and players who create, share, and evolve thousands of new, experimental forms of gameplay—forms that are impossible in closed ecosystems (like Nintendo's) or on offline-only devices.
    
5. **Define Best Practices:** To have our open-source codebase become the go-to standard for how to build secure, low-latency, and scalable IoT infrastructures for real-time human-computer interaction.
    

### 2. Market Landscape

We are not competing with the Steam Deck or Nintendo Switch. We are creating an entirely new, developer-first category that leverages the weaknesses of existing markets.

|**Market**|**Existing Problem**|**The "Crisp" Solution**|
|---|---|---|
|**Closed Consoles** (Switch)|A "walled garden." Innovation is throttled by SDK costs, content curation, and a focus on AAA titles.|**Radical Openness.** Anyone can create and publish. No permission needed. The platform is the community.|
|**Indie Handhelds** (Playdate)|Beautiful but expensive, "boutique" devices. Their focus is often on offline, single-player, retro-style experiences.|**Accessibility & Connection.** "Crisp" runs on affordable, mass-produced hardware (M5). Our core feature is _the network_.|
|**DIY Platforms** (Arduino/Pi)|High friction. To make a game, you must be an expert hardware engineer, C++ programmer, and network specialist.|**Focus & Simplicity.** "Crisp" is an _interactive platform_, not a general-purpose computer. Our SDK handles the hard parts.|

Our Unique Selling Proposition (USP):

"Crisp" is the "Linux for connected handhelds." It is the platform for the breakthrough ideas that are too "weird" for Nintendo, too "networked" for Playdate, and too "complex" to build from scratch on Arduino.

### 3. Scope (The Vision for v1.0)

**What "Crisp" _Is_ (In Scope):**

- **"Crisp Framework" (The Device):** A lightweight, high-performance **C++ framework** (for the Arduino/ESP-IDF toolchain) that runs on the M5 (and other ESP32s). It manages the game loop, graphics, input, and network connection.
    
- **"Crisp SDK" (The Tools):**
    
    - The **C++ Library** for the device (`crisp::connect()`, `crisp::sync_state()`).
        
    - The **JavaScript/TypeScript Library** for the web client (`new CrispClient()`).
        
- **"Crisp Hub" (The Backend):** An open-source, reference backend server written in **Node.js/TypeScript** (or Python). It uses WebSockets to manage rooms, matchmaking, and real-time state synchronization. It's designed for anyone to self-host.
    
- **The Web Client (Equal Platform):** A web-based "Crisp" player is a _first-class citizen_. A game session can be joined by any combination of physical "Crisp" devices and web browsers.
    

**What "Crisp" _Is Not_ (Out of Scope):**
    
- **A Hardware Company:** We do not sell consoles. We create the software that can live inside any compatible hardware.
    
- **A Centralized App Store:** We are not a gatekeeper. We create protocols. Anyone can build their own game "browser" or "store" for "Crisp" games.
    

### 4. Use Cases (Ecosystem Scenarios)

**Scenario 1: The Developer**

- **Actor:** A university student with a game idea.
    
- **Goal:** To create a novel, 4-player, real-time party game.
    
- **How it works:**
    
    1. She downloads the "Crisp C++ SDK" for her M5Stick.
        
    2. She writes her game logic in C++, using the simple `crisp::sync_state()` function to share player positions.
        
    3. She clones the "Crisp Hub" (Node.js) template, adds one line of custom logic, and deploys it for free.
        
    4. She shares her game's source code and the Hub URL. In one weekend, she has created and published a globally-available, cross-platform multiplayer game. _This is the technological breakthrough._
        

**Scenario 2: The Player (The User)**

- **Actor:** A group of friends at a party.
    
- **Goal:** To play a game together.
    
- **How it works:**
    
    1. One friend loads the game on their M5 device. The screen shows a Room Code: `WXYZ`.
        
    2. The second friend opens `crisp.play` (a community-run portal) on their phone's web browser and types `WXYZ`.
        
    3. The third friend does the same on their laptop.
        
    4. All three (M5 device, phone, laptop) are instantly in the same game session, playing together in real-time.
        

**Scenario 3: The Educator (The Mentor)**

- **Actor:** A professor teaching "Networked Systems" or "Game Design."
    
- **Goal:** To give students a tangible, fun platform for learning complex concepts.
    
- **How it works:**
    
    1. The professor uses "Crisp" as the core curriculum.
        
    2. Students write game logic in **C++** (learning about performance, memory, and game loops).
        
    3. They then modify the **Node.js/TypeScript** backend (learning about asynchronous programming, WebSockets, and database state).
        
    4. They graduate having built a _real, functional, cross-platform application_, not just a command-line chat app.
        

### 5. User Stories

- **As a** creator, **I want** to write my game in performant **C++** on my device, **so that** I can have fast graphics and responsive controls.
    
- **As a** backend developer, **I want** a clear and simple **Node.js/TypeScript** (or Python) reference server, **so that** I can easily self-host and customize the logic for my game.
    
- **As a** player, **I want** to join a game from my web browser, **so that** I can play with my friends even if I don't own an M5 device.
    
- **As a** community member, **I want** the entire stack (device framework, backend hub, web client) to be open-source, **so that** I can learn from it, contribute to it, and trust that it will remain free.
    
- **As a** hardware enthusiast, **I want** to easily port the "Crisp C++ Framework" to a new, custom-built ESP32 device, **so that** I can create and share my own unique console designs.