graph TD
    subgraph Клиенты
        Uni[ЛК ВУЗа]
        Stu[ЛК Студента]
        HR[Портал HR]
    end

    subgraph Edge_Security
        GW[API Gateway (Nginx/Traefik)]
        RL[Rate Limiter & WAF]
    end

    subgraph Микросервисы (Go)
        Auth[Auth & RBAC Service]
        Dip[Diploma CRUD & Sign Service]
        QR[QR Token & TTL Service]
        Ver[Verification & Bulk API]
        Audit[Audit & Webhook Service]
    end

    subgraph Data_Layer
        PG[(PostgreSQL 15+)]
        RD[(Redis 7+ Cluster)]
        KF[Apache Kafka 3.x]
    end

    subgraph External_Sec
        KMS[HashiCorp Vault / Yandex KMS]
        FRDO[ФИС ФРДО Adapter (Stub)]
    end

    Uni & Stu & HR --> GW
    GW --> RL
    RL --> Auth & Dip & QR & Ver & Audit
    Auth --> PG & RD
    Dip --> PG & KMS & KF
    QR --> RD & PG
    Ver --> PG & RD & KF
    Audit --> KF
    KF --> Audit
    Dip -.-> FRDO
    KMS -.-> Dip & Auth