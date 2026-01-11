# Event Planning Operating System (SBF-OS)

A backend-driven operating system designed to manage and optimize the planning of complex, high-stakes events. With built-in prioritization, accountability, and auditability, **SBF-OS** streamlines the event planning process to ensure smoother operations and successful outcomes.

**Live Demo:** [Event Planning OS](https://event-planning-os.onrender.com/)  
**Status:** MVP / In Development  

---

## 📌 Problem

Large-scale event planning entails handling hundreds of tasks distributed across multiple participants—volunteers, sponsors, vendors, and logistics teams. Traditionally managed using spreadsheets and fragmented communication tools, these methods often result in:

- **Overlooked high-priority tasks:** Deadlines often slip through the cracks.
- **No ownership:** Silent task failures arise without clarity on who's responsible.
- **Time wasted triaging:** Identifying critical priorities becomes a continual challenge.

These limitations elevate risk, reduce accountability, and impede efficiency.

---

## 🚀 Solution

**SBF-OS** (Event Planning OS) takes an **opinionated and structured approach** to address these challenges. It minimizes *time-to-triage* and maximizes operational effectiveness with smart tools like risk-based prioritization, detailed audit logs, and AI-assisted task breakdowns.

Instead of users wondering, “*What should we focus on next?*”—this system makes the decision for you.

---

## ✨ Key Features

### Risk-Based Task Prioritization 📋
Tasks are automatically triaged and ranked by a computed **Risk Score**. This methodology considers:
- **Overdue tasks:** Prioritizing work that missed its deadline.
- **Upcoming deadlines:** Highlighting tasks closer to their due date.
- **Stale tasks:** Tasks without updates for more than 7 days.
- **Blocked tasks:** Incorporating priority weights for a real-time urgency evaluation.

By eliminating manual sorting, this ensures teams focus on the most critical tasks immediately.

---

### Transactional Audit Logs 📝
Empowered with **granular traceability**, the application's audit system ensures that **what happened**, **when it happened**, and **who initiated it** is always recorded:
- **Immutable events:** Every task update records the transactional JSON diff.
- **Guaranteed rollback:** Any failure in logging automatically negates the update.
- **Comprehensive history:** Systematic logs help track changes over time.

This commitment to auditability safeguards against silent data loss and fosters accountability.

---

### AI-Assisted Task Breakdown 🤖
Leverage the capabilities of **Google Gemini (2.5-Flash)** to transform high-level to-do items into detailed, actionable subtasks based on deterministic planning requirements. Key highlights include:
- JSON-only structured subtasks for machine-readability.
- Predictable behavior, ensuring logical task breakdowns.
- Graceful fallback when an API key isn’t available, without degrading performance.

Rather than replacing planning with automation, AI augments the decision-making pipeline.

---

## 🛠️ Tech Stack

The **SBF-OS** architecture is designed for performance, resilience, and developer-friendliness.

### Backend
- **Language:** [Go (Golang)](https://go.dev/)
- **Routing Library:** [chi](https://github.com/go-chi/chi)
- **Database:** PostgreSQL, with UUID generation and cryptographic extensions via `pgcrypto`.
- **Data Access:** [sqlc](https://sqlc.dev/) (ensures type-safe Go code from native SQL queries).
- **Migrations:** [Goose](https://github.com/pressly/goose) (lightweight database migrations).

### Frontend
- **Rendering:** Server-side rendering with Go `html/template`.
- **Styling Framework:** [PicoCSS](https://picocss.com/) for lightweight, elegant, and responsive layouts.

---

## 🌟 Language Composition

| Language | Percentage |
|----------|------------|
| Go       | 60.9%      |
| HTML     | 39.1%      |

The backend is primarily implemented in **Go** for its concurrency model and performance, while **HTML** powers the frontend for seamless server-rendered views.

---

## 🛠️ Local Setup

### Prerequisites
Before running the application locally, ensure you have:
1. Go 1.25+ installed.
2. PostgreSQL database running locally or in the cloud.
3. *(Optional)* [Google Gemini](https://ai.google/tools/) API key for enabling AI-assisted task breakdown functionality.

### Steps to Clone and Run

1. **Clone the Repository**
   ```bash
   git clone https://github.com/navyaalva/event-planning-os.git
   cd event-planning-os
   ```

2. **Setup Environment Variables**
   Rename `.env.example` to `.env`, then update configurations in `.env` as needed.

3. **Install Dependencies**
   ```bash
   go mod tidy
   ```

4. **Run Database Migrations**
   Apply the defined database schema:
   ```bash
   goose up
   ```

5. **Run the Application**
   ```bash
   go run main.go
   ```

6. **Access the Application**
   - Navigate to **[http://localhost:8080](http://localhost:8080)** in your browser to view the app.

---

## 📢 Contributing

Interested in contributing to **SBF-OS**? We'd love your help! Here's how you can get started:
1. [Fork the repo](https://github.com/navyaalva/event-planning-os/fork).
2. Create a new branch for your feature e.g., `feature/login-improvement`.
3. Submit a Pull Request and describe the changes made.

---

## 📜 License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.

---

## 🌐 Code of Conduct

Please read the [Code of Conduct](CODE_OF_CONDUCT.md) to understand expected behavior while contributing or interacting with this project.

---

Built with ❤️ by [Navya Alva](https://github.com/navyaalva) and contributors.
