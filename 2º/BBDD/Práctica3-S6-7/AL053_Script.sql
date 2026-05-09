-- Generado por Oracle SQL Developer Data Modeler 24.3.1.347.1153
--   en:        2026-05-09 21:12:07 CEST
--   sitio:      Oracle Database 11g
--   tipo:      Oracle Database 11g



-- predefined type, no DDL - MDSYS.SDO_GEOMETRY

-- predefined type, no DDL - XMLTYPE

CREATE TABLE al053_causa (
    al053_relacion_id  INTEGER NOT NULL,
    al053_relacion_id1 INTEGER NOT NULL,
    importancia        VARCHAR2(20) NOT NULL
);

ALTER TABLE al053_causa ADD CONSTRAINT causa_pk PRIMARY KEY ( al053_relacion_id,
                                                              al053_relacion_id1 );

CREATE TABLE al053_centro (
    id           INTEGER NOT NULL,
    especialidad VARCHAR2(50) NOT NULL
);

ALTER TABLE al053_centro ADD CONSTRAINT al053_centro_pk PRIMARY KEY ( id );

CREATE TABLE al053_crea (
    al053_persona_id INTEGER NOT NULL,
    al053_obra_id    INTEGER NOT NULL
);

ALTER TABLE al053_crea ADD CONSTRAINT crea_pk PRIMARY KEY ( al053_persona_id,
                                                            al053_obra_id );

CREATE TABLE al053_empresa (
    id         INTEGER NOT NULL,
    dedicacion VARCHAR2(50) NOT NULL
);

ALTER TABLE al053_empresa ADD CONSTRAINT al053_empresa_pk PRIMARY KEY ( id );

CREATE TABLE al053_formacion (
    id              INTEGER NOT NULL,
    al053_centro_id INTEGER NOT NULL
);

ALTER TABLE al053_formacion ADD CONSTRAINT al053_formacion_pk PRIMARY KEY ( id );

CREATE TABLE al053_institucion (
    id             INTEGER NOT NULL,
    nombre         VARCHAR2(100) NOT NULL,
    creacion       DATE,
    al053_lugar_id INTEGER NOT NULL,
    tipo_inst      VARCHAR2(50) NOT NULL
);

ALTER TABLE al053_institucion
    ADD CONSTRAINT ch_inh_al053_institucion
        CHECK ( tipo_inst IN ( 'AL053_CENTRO', 'AL053_EMPRESA', 'AL053_INSTITUCION' ) );

ALTER TABLE al053_institucion ADD CONSTRAINT al053_institucion_pk PRIMARY KEY ( id );

CREATE TABLE al053_laboral (
    id               INTEGER NOT NULL,
    al053_obra_id    INTEGER NOT NULL,
    al053_empresa_id INTEGER NOT NULL
);

ALTER TABLE al053_laboral ADD CONSTRAINT al053_laboral_pk PRIMARY KEY ( id );

CREATE TABLE al053_lugar (
    id        INTEGER NOT NULL,
    direccion VARCHAR2(150),
    poblacion VARCHAR2(100),
    pais      VARCHAR2(50),
    latitud   NUMBER NOT NULL,
    longitud  NUMBER NOT NULL
);

ALTER TABLE al053_lugar ADD CONSTRAINT al053_lugar_pk PRIMARY KEY ( id );

CREATE TABLE al053_obra (
    id             INTEGER NOT NULL,
    titulo         VARCHAR2(150) NOT NULL,
    creacion       DATE NOT NULL,
    descripcion    VARCHAR2(500),
    tipo           VARCHAR2(50) NOT NULL,
    al053_lugar_id INTEGER NOT NULL
);

ALTER TABLE al053_obra ADD CONSTRAINT al053_obra_pk PRIMARY KEY ( id );

CREATE TABLE al053_persona (
    id             INTEGER NOT NULL,
    nombre         VARCHAR2(50) NOT NULL,
    apellidos      VARCHAR2(100) NOT NULL,
    nacimiento     DATE,
    fallecimiento  DATE,
    al053_lugar_id INTEGER NOT NULL
);

ALTER TABLE al053_persona ADD CONSTRAINT al053_persona_pk PRIMARY KEY ( id );

CREATE TABLE al053_personal (
    id INTEGER NOT NULL
);

ALTER TABLE al053_personal ADD CONSTRAINT al053_personal_pk PRIMARY KEY ( id );

CREATE TABLE al053_relacion (
    id                INTEGER NOT NULL,
    subtipo           VARCHAR2(50) NOT NULL,
    desde             DATE NOT NULL,
    hasta             DATE,
    al053_persona_id  INTEGER NOT NULL,
    al053_persona_id1 INTEGER NOT NULL,
    al053_lugar_id    INTEGER NOT NULL
);

ALTER TABLE al053_relacion
    ADD CONSTRAINT ch_inh_al053_relacion
        CHECK ( subtipo IN ( 'AL053_FORMACION', 'AL053_LABORAL', 'AL053_PERSONAL', 'AL053_RELACION' ) );

ALTER TABLE al053_relacion ADD CONSTRAINT al053_relacion_pk PRIMARY KEY ( id );

ALTER TABLE al053_centro
    ADD CONSTRAINT al053_centro_inst_fk FOREIGN KEY ( id )
        REFERENCES al053_institucion ( id );

ALTER TABLE al053_crea
    ADD CONSTRAINT al053_crea_obra_fk FOREIGN KEY ( al053_obra_id )
        REFERENCES al053_obra ( id );

ALTER TABLE al053_empresa
    ADD CONSTRAINT al053_emp_inst_fk FOREIGN KEY ( id )
        REFERENCES al053_institucion ( id );

ALTER TABLE al053_formacion
    ADD CONSTRAINT al053_form_centro_fk FOREIGN KEY ( al053_centro_id )
        REFERENCES al053_centro ( id );

ALTER TABLE al053_formacion
    ADD CONSTRAINT al053_form_rel_fk FOREIGN KEY ( id )
        REFERENCES al053_relacion ( id );

ALTER TABLE al053_institucion
    ADD CONSTRAINT al053_inst_lugar_fk FOREIGN KEY ( al053_lugar_id )
        REFERENCES al053_lugar ( id );

ALTER TABLE al053_laboral
    ADD CONSTRAINT al053_lab_emp_fk FOREIGN KEY ( al053_empresa_id )
        REFERENCES al053_empresa ( id );

ALTER TABLE al053_laboral
    ADD CONSTRAINT al053_lab_obra_fk FOREIGN KEY ( al053_obra_id )
        REFERENCES al053_obra ( id );

ALTER TABLE al053_laboral
    ADD CONSTRAINT al053_lab_rel_fk FOREIGN KEY ( id )
        REFERENCES al053_relacion ( id );

ALTER TABLE al053_obra
    ADD CONSTRAINT al053_lug_obra_fk FOREIGN KEY ( al053_lugar_id )
        REFERENCES al053_lugar ( id );

ALTER TABLE al053_crea
    ADD CONSTRAINT al053_pers_causa_fk FOREIGN KEY ( al053_persona_id )
        REFERENCES al053_persona ( id );

ALTER TABLE al053_personal
    ADD CONSTRAINT al053_pers_rel_fk FOREIGN KEY ( id )
        REFERENCES al053_relacion ( id );

ALTER TABLE al053_persona
    ADD CONSTRAINT al053_persona_al053_lugar_fk FOREIGN KEY ( al053_lugar_id )
        REFERENCES al053_lugar ( id );

ALTER TABLE al053_relacion
    ADD CONSTRAINT al053_rel_lug_fk FOREIGN KEY ( al053_lugar_id )
        REFERENCES al053_lugar ( id );

ALTER TABLE al053_relacion
    ADD CONSTRAINT al053_rel_pers1_fk FOREIGN KEY ( al053_persona_id1 )
        REFERENCES al053_persona ( id );

ALTER TABLE al053_relacion
    ADD CONSTRAINT al053_rel_pers2_fk FOREIGN KEY ( al053_persona_id )
        REFERENCES al053_persona ( id );

ALTER TABLE al053_causa
    ADD CONSTRAINT causa_al053_relacion_fk FOREIGN KEY ( al053_relacion_id )
        REFERENCES al053_relacion ( id );

ALTER TABLE al053_causa
    ADD CONSTRAINT causa_al053_relacion_fkv1 FOREIGN KEY ( al053_relacion_id1 )
        REFERENCES al053_relacion ( id );

CREATE OR REPLACE TRIGGER arc_fkarc_1_al053_centro BEFORE
    INSERT OR UPDATE OF id ON al053_centro
    FOR EACH ROW
DECLARE
    d VARCHAR2(50);
BEGIN
    SELECT
        a.tipo_inst
    INTO d
    FROM
        al053_institucion a
    WHERE
        a.id = :new.id;

    IF ( d IS NULL
         OR d <> 'AL053_CENTRO' ) THEN
        raise_application_error(-20223, 'FK AL053_CENTRO_INST_FK in Table AL053_CENTRO violates Arc constraint on Table AL053_INSTITUCION - discriminator column tipo_inst doesn''t have value ''AL053_CENTRO'''
        );
    END IF;

EXCEPTION
    WHEN no_data_found THEN
        NULL;
    WHEN OTHERS THEN
        RAISE;
END;
/

CREATE OR REPLACE TRIGGER arc_fkarc_1_al053_empresa BEFORE
    INSERT OR UPDATE OF id ON al053_empresa
    FOR EACH ROW
DECLARE
    d VARCHAR2(50);
BEGIN
    SELECT
        a.tipo_inst
    INTO d
    FROM
        al053_institucion a
    WHERE
        a.id = :new.id;

    IF ( d IS NULL
         OR d <> 'AL053_EMPRESA' ) THEN
        raise_application_error(-20223, 'FK AL053_EMP_INST_FK in Table AL053_EMPRESA violates Arc constraint on Table AL053_INSTITUCION - discriminator column tipo_inst doesn''t have value ''AL053_EMPRESA'''
        );
    END IF;

EXCEPTION
    WHEN no_data_found THEN
        NULL;
    WHEN OTHERS THEN
        RAISE;
END;
/

CREATE OR REPLACE TRIGGER arc_fkarc_2_al053_personal BEFORE
    INSERT OR UPDATE OF id ON al053_personal
    FOR EACH ROW
DECLARE
    d VARCHAR2(50);
BEGIN
    SELECT
        a.subtipo
    INTO d
    FROM
        al053_relacion a
    WHERE
        a.id = :new.id;

    IF ( d IS NULL
         OR d <> 'AL053_PERSONAL' ) THEN
        raise_application_error(-20223, 'FK AL053_PERS_REL_FK in Table AL053_PERSONAL violates Arc constraint on Table AL053_RELACION - discriminator column subtipo doesn''t have value ''AL053_PERSONAL'''
        );
    END IF;

EXCEPTION
    WHEN no_data_found THEN
        NULL;
    WHEN OTHERS THEN
        RAISE;
END;
/

CREATE OR REPLACE TRIGGER arc_fkarc_2_al053_formacion BEFORE
    INSERT OR UPDATE OF id ON al053_formacion
    FOR EACH ROW
DECLARE
    d VARCHAR2(50);
BEGIN
    SELECT
        a.subtipo
    INTO d
    FROM
        al053_relacion a
    WHERE
        a.id = :new.id;

    IF ( d IS NULL
         OR d <> 'AL053_FORMACION' ) THEN
        raise_application_error(-20223, 'FK AL053_FORM_REL_FK in Table AL053_FORMACION violates Arc constraint on Table AL053_RELACION - discriminator column subtipo doesn''t have value ''AL053_FORMACION'''
        );
    END IF;

EXCEPTION
    WHEN no_data_found THEN
        NULL;
    WHEN OTHERS THEN
        RAISE;
END;
/

CREATE OR REPLACE TRIGGER arc_fkarc_2_al053_laboral BEFORE
    INSERT OR UPDATE OF id ON al053_laboral
    FOR EACH ROW
DECLARE
    d VARCHAR2(50);
BEGIN
    SELECT
        a.subtipo
    INTO d
    FROM
        al053_relacion a
    WHERE
        a.id = :new.id;

    IF ( d IS NULL
         OR d <> 'AL053_LABORAL' ) THEN
        raise_application_error(-20223, 'FK AL053_LAB_REL_FK in Table AL053_LABORAL violates Arc constraint on Table AL053_RELACION - discriminator column subtipo doesn''t have value ''AL053_LABORAL'''
        );
    END IF;

EXCEPTION
    WHEN no_data_found THEN
        NULL;
    WHEN OTHERS THEN
        RAISE;
END;
/



-- Informe de Resumen de Oracle SQL Developer Data Modeler: 
-- 
-- CREATE TABLE                            12
-- CREATE INDEX                             0
-- ALTER TABLE                             32
-- CREATE VIEW                              0
-- ALTER VIEW                               0
-- CREATE PACKAGE                           0
-- CREATE PACKAGE BODY                      0
-- CREATE PROCEDURE                         0
-- CREATE FUNCTION                          0
-- CREATE TRIGGER                           5
-- ALTER TRIGGER                            0
-- CREATE COLLECTION TYPE                   0
-- CREATE STRUCTURED TYPE                   0
-- CREATE STRUCTURED TYPE BODY              0
-- CREATE CLUSTER                           0
-- CREATE CONTEXT                           0
-- CREATE DATABASE                          0
-- CREATE DIMENSION                         0
-- CREATE DIRECTORY                         0
-- CREATE DISK GROUP                        0
-- CREATE ROLE                              0
-- CREATE ROLLBACK SEGMENT                  0
-- CREATE SEQUENCE                          0
-- CREATE MATERIALIZED VIEW                 0
-- CREATE MATERIALIZED VIEW LOG             0
-- CREATE SYNONYM                           0
-- CREATE TABLESPACE                        0
-- CREATE USER                              0
-- 
-- DROP TABLESPACE                          0
-- DROP DATABASE                            0
-- 
-- REDACTION POLICY                         0
-- 
-- ORDS DROP SCHEMA                         0
-- ORDS ENABLE SCHEMA                       0
-- ORDS ENABLE OBJECT                       0
-- 
-- ERRORS                                   0
-- WARNINGS                                 0
