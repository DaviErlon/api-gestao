-- migration.sql
-- Execute com: psql -U <usuario> -d <banco> -f migration.sql

CREATE TABLE IF NOT EXISTS empresas (
    id           SERIAL PRIMARY KEY,
    name         TEXT NOT NULL
);

INSERT INTO empresas (id, name) VALUES
(1, 'Alpha'),
(2, 'Beta'),
(3, 'Gamma'),
(4, 'Delta')
ON CONFLICT (id) DO NOTHING;


CREATE TABLE IF NOT EXISTS users (
    id       SERIAL PRIMARY KEY,
    name     TEXT NOT NULL,
    login    TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    empresa_id   INT,
    
    FOREIGN KEY (empresa_id) REFERENCES empresas(id) ON DELETE CASCADE
);

INSERT INTO users (id, name, login, password, empresa_id) VALUES
(1, 'Admin', 'admin@user', 'user@user', NULL),
(2, 'Filipe', 'filipe@user', 'user@user', 4),
(3, 'Daniel', 'daniel@user', 'user@user', 4),
(4, 'Francisco', 'francisco@user', 'user@user', 4),
(5, 'Rian', 'rian@user', 'user@user', 4),
(6, 'Davi', 'davi@user', 'user@user', 4)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS ciclos (
    id           SERIAL PRIMARY KEY,
    rodada       INT           NOT NULL UNIQUE DEFAULT 0,

    saldo        NUMERIC(15,2) NOT NULL DEFAULT 0,
    divida       NUMERIC(15,2) NOT NULL DEFAULT 0,
    juros        NUMERIC(5,2)  DEFAULT 0, 
    clientes     INT           NOT NULL DEFAULT 0,
    market_share NUMERIC(5,2)  NOT NULL DEFAULT 0,
    tech         INT           NOT NULL DEFAULT 0,
    reputacao    INT           NOT NULL DEFAULT 0,
    seguranca    INT           NOT NULL DEFAULT 0,
    nps          INT           NOT NULL DEFAULT 0,
    valuation    NUMERIC(15,2) NOT NULL DEFAULT 0,

    empresa_id   INT NOT NULL,
    
    FOREIGN KEY (empresa_id) REFERENCES empresas(id) ON DELETE CASCADE,

    UNIQUE (empresa_id, rodada)
);

CREATE TABLE IF NOT EXISTS decisoes (
    id          SERIAL PRIMARY KEY,
    marketing   NUMERIC(15,2) NOT NULL DEFAULT 0,
    ped         NUMERIC(15,2) NOT NULL DEFAULT 0,
    suporte     NUMERIC(15,2) NOT NULL DEFAULT 0,
    seguranca   NUMERIC(15,2) NOT NULL DEFAULT 0,
    expansao    INT           NOT NULL DEFAULT 0,
    empresa_id  INT NOT NULL,
    ciclo_id    INT NOT NULL,
    
    FOREIGN KEY (empresa_id) REFERENCES empresas(id) ON DELETE CASCADE,
    FOREIGN KEY (ciclo_id) REFERENCES ciclos(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS posts (
    id        SERIAL PRIMARY KEY,
    content   TEXT NOT NULL,
    author_id INT  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pin       BOOLEAN NOT NULL DEFAULT FALSE,
    posted_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS likes (
    usuario_id INT NOT NULL,
    post_id INT NOT NULL,
    PRIMARY KEY (usuario_id, post_id),
    FOREIGN KEY (usuario_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS reunioes (
    id        SERIAL PRIMARY KEY,
    ciclo_id  INT NOT NULL,
    author_id INT NOT NULL,
    titulo    TEXT NOT NULL,
    descricao TEXT,

    inicio    TIMESTAMP NOT NULL,
    duracao   INTERVAL NOT NULL,

    FOREIGN KEY (ciclo_id) REFERENCES ciclos(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);