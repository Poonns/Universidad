--eliminar tablas
DROP TABLE RESULTADO;
DROP TABLE UTILIZA;
DROP TABLE MUESTRA;
DROP TABLE REACTIVO;
DROP TABLE EQUIPO;
DROP TABLE PRUEBA;
DROP TABLE CENTRO_PROCEDENCIA;
DROP TABLE TECNICO;
DROP TABLE MEDICO;
DROP TABLE PACIENTE;
DROP TABLE PERSONA;

--tablas
CREATE TABLE PERSONA (
    dni VARCHAR2(9),
    nombre VARCHAR2(50) NOT NULL,
    apellidos VARCHAR2(100) NOT NULL,
    CONSTRAINT pk_persona PRIMARY KEY (dni)
);

CREATE TABLE PACIENTE (
    dni VARCHAR2(9),
    telefono VARCHAR2(15) NOT NULL,
    fechaNac DATE NOT NULL,
    CONSTRAINT pk_paciente PRIMARY KEY (dni),
    CONSTRAINT fk_paciente_persona FOREIGN KEY (dni) REFERENCES PERSONA(dni)
);

CREATE TABLE MEDICO (
    dni VARCHAR2(9),
    idMedico VARCHAR2(20) NOT NULL,
    especialidad VARCHAR2(50) NOT NULL,
    CONSTRAINT pk_medico PRIMARY KEY (dni),
    CONSTRAINT uk_medico_id UNIQUE (idMedico),
    CONSTRAINT fk_medico_persona FOREIGN KEY (dni) REFERENCES PERSONA(dni)
);

CREATE TABLE TECNICO (
    dni VARCHAR2(9),
    funcion VARCHAR2(50) NOT NULL,
    CONSTRAINT pk_tecnico PRIMARY KEY (dni),
    CONSTRAINT fk_tecnico_persona FOREIGN KEY (dni) REFERENCES PERSONA(dni)
);

CREATE TABLE CENTRO_PROCEDENCIA (
    idCentro NUMBER,
    nombre VARCHAR2(100) NOT NULL,
    direccion VARCHAR2(200) NOT NULL,
    telefono VARCHAR2(15) NOT NULL,
    CONSTRAINT pk_centro PRIMARY KEY (idCentro)
);

CREATE TABLE PRUEBA (
    idPrueba NUMBER,
    nombre VARCHAR2(100) NOT NULL,
    valorMin NUMBER(10,2) NOT NULL,
    valorMax NUMBER(10,2) NOT NULL,
    unidadMedida VARCHAR2(20) NOT NULL,
    CONSTRAINT pk_prueba PRIMARY KEY (idPrueba)
);

CREATE TABLE EQUIPO (
    idEquipo NUMBER,
    marca VARCHAR2(50) NOT NULL,
    modelo VARCHAR2(50) NOT NULL,
    estado VARCHAR2(20) NOT NULL,
    CONSTRAINT pk_equipo PRIMARY KEY (idEquipo)
);

CREATE TABLE REACTIVO (
    idReactivo NUMBER,
    nombre VARCHAR2(100) NOT NULL,
    lote VARCHAR2(50) NOT NULL,
    fechaCaducidad DATE NOT NULL,
    CONSTRAINT pk_reactivo PRIMARY KEY (idReactivo)
);

CREATE TABLE UTILIZA (
    idEquipo NUMBER,
    idReactivo NUMBER,
    cantidadConsumida NUMBER(10,2) NOT NULL,
    CONSTRAINT pk_utiliza PRIMARY KEY (idEquipo, idReactivo),
    CONSTRAINT fk_utiliza_equipo FOREIGN KEY (idEquipo) REFERENCES EQUIPO(idEquipo),
    CONSTRAINT fk_utiliza_reactivo FOREIGN KEY (idReactivo) REFERENCES REACTIVO(idReactivo)
);

CREATE TABLE MUESTRA (
    idMuestra NUMBER,
    idCentro NUMBER,
    dniPaciente VARCHAR2(9) NOT NULL,
    dniMedico VARCHAR2(9) NOT NULL,
    fechaExtraccion DATE NOT NULL,
    tipo VARCHAR2(50) NOT NULL,
    estado VARCHAR2(20) NOT NULL,
    CONSTRAINT pk_muestra PRIMARY KEY (idMuestra),
    CONSTRAINT fk_muestra_centro FOREIGN KEY (idCentro) REFERENCES CENTRO_PROCEDENCIA(idCentro),
    CONSTRAINT fk_muestra_paciente FOREIGN KEY (dniPaciente) REFERENCES PACIENTE(dni),
    CONSTRAINT fk_muestra_medico FOREIGN KEY (dniMedico) REFERENCES MEDICO(dni)
);

CREATE TABLE RESULTADO (
    idMuestra NUMBER,
    idPrueba NUMBER,
    dniTecnico VARCHAR2(9) NOT NULL,
    idEquipo NUMBER NOT NULL,
    valorObtenido NUMBER(10,2) NOT NULL,
    estadoAnomalo NUMBER(1),
    fechaHora DATE NOT NULL,
    observaciones VARCHAR2(255),
    CONSTRAINT pk_resultado PRIMARY KEY (idMuestra, idPrueba),
    CONSTRAINT fk_resultado_muestra FOREIGN KEY (idMuestra) REFERENCES MUESTRA(idMuestra),
    CONSTRAINT fk_resultado_prueba FOREIGN KEY (idPrueba) REFERENCES PRUEBA(idPrueba),
    CONSTRAINT fk_resultado_tecnico FOREIGN KEY (dniTecnico) REFERENCES TECNICO(dni),
    CONSTRAINT fk_resultado_equipo FOREIGN KEY (idEquipo) REFERENCES EQUIPO(idEquipo)
);
--indices
CREATE INDEX idx_muestra_estado ON MUESTRA(estado);
CREATE INDEX idx_resultado_fechahora ON RESULTADO(fechaHora);