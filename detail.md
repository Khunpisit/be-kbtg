# System Detail Documentation

เอกสารนี้อธิบายลำดับการทำงาน (Sequence) ของการสมัครสมาชิกและเข้าสู่ระบบ พร้อมทั้ง ER Diagram (Mermaid) ของตารางผู้ใช้ (users)

## 1. Sequence Diagram: สมัครสมาชิก (Register) & เข้าสู่ระบบ (Login)

```mermaid
sequenceDiagram
    autonumber
    actor U as User
    participant FE as Frontend (Browser / App)
    participant API as Fiber API (HTTP)
    participant SVC as UserService / UseCase
    participant REPO as GORM Repository
    participant DB as SQLite DB (app.db)
    participant JWT as JWT Module

    rect rgb(235, 248, 255)
    note over U,API: Flow: Register
    U->>FE: กรอกอีเมล / รหัสผ่าน / ชื่อ
    FE->>API: POST /api/v1/auth/register (JSON)
    API->>SVC: Validate & Map DTO
    SVC->>REPO: Check existing email
    REPO->>DB: SELECT users WHERE email = ?
    DB-->>REPO: result (empty)
    SVC->>SVC: Hash password (bcrypt/crypto)
    SVC->>REPO: Create User
    REPO->>DB: INSERT INTO users (...)
    DB-->>REPO: inserted row (ID)
    REPO-->>SVC: User entity
    SVC->>JWT: Generate Access Token
    JWT-->>SVC: token (signed)
    SVC-->>API: User + token
    API-->>FE: 201 Created (JSON)
    FE-->>U: แสดงผลสำเร็จ
    end

    rect rgb(255, 245, 230)
    note over U,API: Flow: Login
    U->>FE: กรอกอีเมล / รหัสผ่าน
    FE->>API: POST /api/v1/auth/login
    API->>SVC: Validate DTO
    SVC->>REPO: Find user by email
    REPO->>DB: SELECT * FROM users WHERE email = ?
    DB-->>REPO: user row
    REPO-->>SVC: User entity
    SVC->>SVC: Compare password hash
    alt รหัสผ่านถูกต้อง
        SVC->>JWT: Generate Access Token
        JWT-->>SVC: token
        SVC-->>API: User + token
        API-->>FE: 200 OK (JSON)
        FE-->>U: เข้าสู่ระบบสำเร็จ
    else รหัสผ่านไม่ถูกต้อง
        SVC-->>API: error (401)
        API-->>FE: 401 Unauthorized
        FE-->>U: แจ้งเตือนข้อผิดพลาด
    end
    end
```

## 2. ER Diagram (Mermaid)

```mermaid
erDiagram
    USERS {
        uint id PK "Primary Key"
        string email "Unique, Not Null"
        string password_hash "Hashed password (ไม่ส่งออกใน JSON)"
        string first_name
        string last_name
        string display_name
        string phone
        string avatar_url
        string bio
        datetime created_at
        datetime updated_at
    }
```

หมายเหตุ:
- ตารางที่มีอยู่ในโค้ดตอนนี้เห็นเพียงโครงสร้าง User (models/user.go)
- หากในอนาคตมีตารางอื่น (เช่น POSTS, ROLES, REFRESH_TOKENS) สามารถขยาย ER Diagram ได้โดยเพิ่มความสัมพันธ์ (|o--o{, ||--o{ ฯลฯ)

## 3. แนวคิดการขยาย (Future Extension)
- เพิ่มตาราง refresh_tokens เพื่อรองรับ refresh flow
- เพิ่มตาราง roles / permissions แล้วทำตาราง many-to-many (user_roles)
- จัดทำ index เพิ่มเติม: UNIQUE(email), INDEX(phone)

## 4. การรักษาความปลอดภัย
- Password เก็บเฉพาะ Hash (เช่น bcrypt / Argon2)
- JWT ควรมี expiration (เช่น 15m) และ secret เก็บใน environment variable
- ควรเพิ่ม middleware ตรวจสอบ JWT บนเส้นทางที่ต้องการ authentication

จบเอกสาร
