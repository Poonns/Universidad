/* Ejercicio 1: Escribe una sentencia SQL para crear tu tabla EMPLEADOS. Debes crear la tabla con
todos los atributos indicados y la restricción de clave primaria. */
DROP TABLE EMPLEADOS CASCADE CONSTRAINTS;

CREATE TABLE EMPLEADOS (
    idempleado NUMBER,
    email VARCHAR2(30),
    telefono NUMBER(9),
    trabajo VARCHAR(20),
    salario FLOAT,
    
    CONSTRAINT empleado_pk PRIMARY KEY(idempleado));

/*Ejercicio 2: Escribe una sentencia SQL para insertar en la tabla EMPLEADOS las siguientes
restricciones:
a) Restricción que evite que el email sea incorrecto.
b) Restricción que asegure que los teléfonos de la tabla EMPLEADOS son correctos.
c) Asegúrate que los trabajos que se puedan insertar estén en el siguiente conjunto
(Instructor, Administrativo, Comercial, Vendedor, Veterinario).
d) El salario debe ser superior a 655,20 € (salario mínimo interprofesional).
e) El atributo email debe ser único.
f) El atributo teléfono debe ser único.*/
ALTER TABLE EMPLEADOS
    ADD(
    CONSTRAINT email_check CHECK(email LIKE '%@%'),
    CONSTRAINT telefono_check CHECK(telefono BETWEEN 600000000 AND 999999999),
    CONSTRAINT trabajo_check CHECK(UPPER(trabajo) IN ('INSTRUCTOR', 'ADMINISTRATIVO', 'COMERCIAL', 'VENDEDOR', 'VETERINARIO')),
    CONSTRAINT salario_check CHECK(salario > 655.20),
    CONSTRAINT email_unique UNIQUE(email),
    CONSTRAINT telefono_unique UNIQUE(telefono)
    );

/*Ejercicio 3: Escribe una sentencia SQL para añadir en la tabla CLIENTES las siguientes
restricciones:
a) El email debe ser válido.
b) Los nombres y apellidos se deben almacenar en mayúsculas.*/
ALTER TABLE CLIENTES
    ADD(
        CONSTRAINT email_check CHECK(email LIKE '%@%'),
        CONSTRAINT nombres_apellido_check CHECK(nombre = UPPER(nombre) AND apellido = UPPER(apellido))
    );

/*Ejercicio 4: Añade la restricción de clave ajena en la tabla cursos que impide que el atributo
tipoanimal tome valores distintos de los que aparecen en el atributo tipoanimal de la tabla
ANIMALES.
a) ¿Puedes hacerlo?
b) Si no, ¿Qué error obtienes y por qué? Di cómo solucionarías el problema*/
ALTER TABLE CURSOS
    ADD (CONSTRAINT tipoanimal_fk FOREIGN KEY(tipoanimal)
        REFERENCES ANIMALES(tipoanimal));
-- Depende, hay que tener que tener en cuenta que la referencia debe de ser un PK o que sea unica

/*Ejercicio 5: Añade el resto de restricciones que aparecen en el enunciado sobre la tabla CURSOS.*/
ALTER TABLE CURSOS
    ADD(
        CONSTRAINT max_num_check CHECK(max_num BETWEEN 10 AND 30),
        CONSTRAINT date_check CHECK(EXTRACT(MONTH FROM FECHAHORA) BETWEEN 7 AND 9),
        CONSTRAINT hora_check CHECK(EXTRACT(HOUR FROM FECHAHORA) >= 15),
        CONSTRAINT instructor_fk FOREIGN KEY(instructor)
            REFERENCES EMPLEADOS(idempleado)
    );

/*Ejercicio 6: Crea la tabla MASCOTAS utilizando la información proporcionada en el enunciado de
la práctica.*/
DROP TABLE MASCOTAS CASCADE CONSTRAINTS;

CREATE TABLE MASCOTAS (
    idmascota NUMBER,
    idcliente NUMBER,

    CONSTRAINT idmascota_pk PRIMARY KEY(idmascota),
    CONSTRAINT idcliente_fk FOREIGN KEY(idcliente)
        REFERENCES CLIENTES(idcliente)
);

/*Ejercicio 7: Crea la tabla CURSOS_CLIENTES utilizando la información proporcionada en el
enunciado de la práctica.*/
DROP TABLE CURSOS_CLIENTES CASCADE CONSTRAINTS;

CREATE TABLE CURSOS_CLIENTES (
    idcurso NUMBER,
    idcliente NUMBER,

    CONSTRAINT curso_cliente_pk PRIMARY KEY(idcurso, idcliente),
    CONSTRAINT curso_fk FOREIGN KEY(idcurso)
        REFERENCES CURSOS(idcurso),
    CONSTRAINT cliente_fk FOREIGN KEY(idcliente)
        REFERENCES CLIENTES(idcliente)
);

/*Ejercicio 9: Crea la tabla ANIMALES utilizando la tabla existente en el esquema LBD2011. ¿Se
copian también las restricciones? Si no, añadir las restricciones oportunas.*/
DROP TABLE ANIMALES CASCADE CONSTRAINTS;

CREATE TABLE ANIMALES (
    tipoanimal CONSTRAINT tipoanimal_pk PRIMARY KEY(tipoanimal),
    nombre CONSTRAINT nombre_not_null NOT NULL
)
    AS SELECT * FROM LBD2011.ANIMALES;

/*Ejercicio 10: Crea las tablas PEDIDOS, PEDIDOS_PRODUCTOS y PRODUCTOS utilizando la
información proporcionada por el enunciado de la práctica. Créalas en el orden que consideres
oportuno.*/
DROP TABLE PEDIDOS CASCADE CONSTRAINTS;
DROP TABLE PEDIDOS_PRODUCTOS CASCADE CONSTRAINTS;
DROP TABLE PRODUCTOS CASCADE CONSTRAINTS;

CREATE TABLE PEDIDOS (
    idpedido NUMBER,
    idcliente NUMBER,
    empleadoventa NUMBER,

    CONSTRAINT idepedido_pk PRIMARY KEY(idpedido),
    CONSTRAINT idcliente_fk FOREIGN KEY(idcliente)
        REFERENCES CLIENTES(idcliente),
    CONSTRAINT empleadoventa_fk FOREIGN KEY(empleadoventa)
        REFERENCES EMPLEADoS(idempleado)
);

CREATE TABLE PRODUCTOS (
    idproducto NUMBER,
    precio_compra FLOAT,
    precio_venta FLOAT,
    stock NUMBER,
    idproveedor NUMBER,
    descripcion VARCHAR2(100) NOT NULL,

    CONSTRAINT idprocuto_pk PRIMARY KEY(idproducto),
    CONSTRAINT precio_compra_check CHECK(precio_compra < precio_venta),
    CONSTRAINT idproveedores_fk FOREIGN KEY(idproveedor)
        REFERENCES PROVEEDORES(idproveedor)
);

CREATE TABLE PEDIDOS_PRODUCTOS (
    idproducto NUMBER,
    idpedido NUMBER,

    CONSTRAINT producto_pedido_pk PRIMARY KEY(idproducto, idpedido),
    CONSTRAINT idproducto_fk FOREIGN KEY(idproducto)
        REFERENCES PRODUCTOS(idproducto),
    CONSTRAINT idpedido_fk FOREIGN KEY(idpedido)
        REFERENCES PEDIDOS(idpedido)
);

/*Ejercicio 12: Para todas las claves ajenas, revisa que claves ajenas no pueden tomar el valor NULL
y añade las restricciones que correspondan.*/
ALTER TABLE MASCOTAS
    MODIFY(idcliente CONSTRAINT idcliente_not_null NOT NULL);

ALTER TABLE CURSOS
    MODIFY(tipoanimal CONSTRAINT tipoanimal_not_null NOT NULL);

ALTER TABLE PEDIDOS
    MODIFY(idcliente CONSTRAINT idcliente_not_null NOT NULL);

ALTER TABLE PRODUCTOS
    MODIFY(idproveedor CONSTRAINT idproveedor_not_null NOT NULL);

/*Ejercicio 13: ¿Sobre qué atributos crearías índices? Di cuáles y por qué. Crea los índices que has
seleccionado.*/
CREATE INDEX idx_idcliente ON MASCOTAS(idcliente);
CREATE INDEX idx_tipoanimal ON CURSOS(tipoanimal);
CREATE INDEX idx_idcliente ON PEDIDOS(idcliente);
CREATE INDEX idx_idproveedor ON PRODUCTOS(idproveedor);

/*Ejercicio 14: Carga los datos de todas las tablas creadas en tu esquema utilizando las tablas base
que utilizamos en la práctica 1 (tablas en el esquema LBD2011). Piensa en el orden correcto de
ejecución de las distintas sentencias.*/
INSERT INTO EMPLEADOS_
    SELECT * FROM LBD2011.EMPLEADOS;

INSERT INTO CLIENTES_
    SELECT * FROM LBD2011.CLIENTES;

INSERT INTO CURSOS_
    SELECT * FROM LBD2011.CURSOS;

INSERT INTO ANIMALES_
    SELECT * FROM LBD2011.ANIMALES;

INSERT INTO CURSOS_CLIENTES_
    SELECT * FROM LBD2011.CURSOS_CLIENTES;

INSERT INTO PROVEEDORES_
    SELECT * FROM LBD2011.PROVEEDORES;

INSERT INTO PRODUCTOS_
    SELECT * FROM LBD2011.PRODUCTOS;

INSERT INTO PEDIDOS_
    SELECT * FROM LBD2011.PEDIDOS;

INSERT INTO PEDIDOS_PRODUCTOS_
    SELECT * FROM LBD2011.PEDIDOS_PRODUCTOS;

INSERT INTO MASCOTAS_
    SELECT * FROM LBD2011.MASCOTAS;

/*Ejercicio 15: Inserta un nuevo cliente en la tabla CLIENTES que Ana María Fernández Morales,
que vive en Plaza del Ayuntamiento, nº 1 de Valencia. Su email amfm@gmail.com y no tiene
teléfono.*/
INSERT INTO CLIENTES(idcliente, nombre, apellido, email, direccionpostal, ciudad, telefono)
VALUES(50, 'ANA MARÍA', 'FERNÁNDEZ MORALES', 'amfm@gmail.com', 'Plaza del Ayuntamiento, nº1', 'Valencia', NULL);

/*Ejercicio 16: El cliente con nombre ‘SARA’ y apellido ‘ALBERTO’ se quiere matricular del curso
con título ‘Como tratar una mascota’. Inserta una tupla en la tabla CURSOS_CLIENTES que dé
cuenta de ello.*/
INSERT INTO CURSOS_CLIENTES(idcurso, idcliente)
    SELECT idcurso, idcliente
    FROM CURSOS c , CLIENTES cli
    WHERE cli.nombre = 'SARA' AND cli.apellido = 'ALBERTO' AND c.titulo = 'Como tratar una mascota';

/*Ejercicio 17: El curso titulado ‘Higiene de gatos especiales’ no se va a impartir por no llegar al
mínimo número de alumnos necesarios. Borra todas las matrículas de alumnos de ese curso.
¿Cuántas matrículas de alumnos se han borrado?*/
DELETE FROM CURSOS_CLIENTES
    WHERE idcliente = (SELECT idcurso FROM CURSOS
        WHERE titulo = 'Higiene de gatos especiales');

/*Ejercicio 18: Queremos hacer una promoción que consiste en rebajar un 10% el precio de venta del
producto/productos más vendido/s. Más vendido significa que aparece con más frecuencia en los
pedidos de la empresa.*/
UPDATE PRODUCTOS
    SET precio_venta = precio_venta * 0.9
    WHERE idproducto IN (SELECT idproducto
        FROM PEDIDOS_PRODUCTOS
        GROUP BY idproducto
        HAVING SUM(cantidad) = (SELECT MAX(SUM(cantidad))
            FROM PEDIDOS PRODUCTOS
            GROUP BY idproducto)
        );

/*Ejercicio 19: Añade dos atributos nuevos a la tabla EMPLEADOS para almacenar información
sobre posibles incentivos salariales por antigüedad. Para ello se deberá crear un campo que refleje el
número de trienios que tiene y otro campo que indique el aumento salarial.*/
ALTER TABLE EMPLEADOS
    ADD(num_trienio NUMBER,
        aumento_salarial FLOAT);

/*Ejercicio 20: Borra todos los campos extras creados en el ejercicio 19, y en el ejemplo 10.*/
ALTER TABLE EMPLEADOS
    DROP (num_trienio, aumento_salarial);

ALTER TABLE CLIENTES
    DROP COLUMN fecha_nac;

/*Ejercicio 21: La cliente ‘Ana María Fernández’ ha ejercido su derecho sobre LOPD y desea ser
borrada de nuestra BD. Realiza los cambios oportunos en los datos almacenados para dar cuenta de
este hecho.*/
DELETE FROM MASCOTAS
    WHERE idcliente = (SELECT idcliente FROM CLIENTES
        WHERE UPPER(nombre) = 'ANA MÁRIA' AND UPPER(apellido) = 'FERNÁNDEZ MORALES');

DELETE FROM PEDIDOS
    WHERE idcliente = (SELECT idcliente FROM CLIENTES
        WHERE UPPER(nombre) = 'ANA MÁRIA' AND UPPER(apellido) = 'FERNÁNDEZ MORALES');

DELETE FROM CURSOS_CLIENTES
    WHERE idcliente = (SELECT idcliente FROM CLIENTES
        WHERE UPPER(nombre) = 'ANA MÁRIA' AND UPPER(apellido) = 'FERNÁNDEZ MORALES');

DELETE FROM PEDIDOS_CLIENTES
    WHERE idpedido = (SELECT idepedido FROM PEDIDO
        WHERE idcliente = (SELECT idcliente FROM CLIENTES
            WHERE UPPER(nombre) = 'ANA MÁRIA' AND UPPER(apellido) = 'FERNÁNDEZ MORALES'));

DELETE FROM CLIENTES
    WHERE UPPER(nombre) = 'ANA MÁRIA' AND UPPER(apellido) = 'FERNÁNDEZ MORALES';

/*Ejercicio 22: Deseamos crear una tabla de estadísticas estáticas que guarde un resumen de la
información sobre los instructores que imparten los cursos así como el saber cuántos han sido, para
cuántos clientes, etc.., en un momento dado.
a) Para ello realizaremos una tabla previa llamada InformacionMatriculaCursos en la que
almacenaremos la información del instructor (identificador, nombre, apellidos y salario), así como
el código del curso impartido, la cantidad de alumnos matriculados por cada curso impartido así
como el total cobrado por cada curso (suponiendo que el total es el producto del precio de cada
curso por el número de alumnos). Esta tabla se debe poblar con datos de la BDs.
Haciendo uso de esa nueva tabla, crearemos una nueva tabla llamada ResumenEstadisticaCursos en
la que almacenaremos: el identificador, nombre, apellido del instructor que impartió el curso así
como su salario, el número total de cursos impartidos, el promedio de alumnos matriculados por
curso, el máximo número de alumnos matriculados en un curso y el mínimo, saber lo que se ha
cobrado en los cursos en promedio, así como el máximo que se cobró por un curso y el mínimo.
b) Reescribe las sentencias anteriores de manera que solo se cree una única tabla llamada
EstadísticasResumen pero que guarde los mismos resultados que ResumenEstadisticaCursos.*/
DROP TABLE InformacionMatriculaCursos CASCADE CONSTRAINTS;

CREATE TABLE InformacionMatriculaCursos AS  
SELECT e.idempleado AS id_instructor, e.nombre, e.apellido, c.idcurso, COUNT(cc.idcliente) AS num_alumnos, (COUNT(cc.idcliente) * c.precio)
FROM EMPLEADOS e JOIN CURSOS c ON e.idempleado = c.instructor JOIN CURSOS_CLIENTES cc ON c.idcurso = cc.idcurso 
GROUP BY e.idempleado, e.nombre, e.apellidos, e.salario, c.idcurso, c.precio;

DROP TABLE ResumenEstadisticaCursos CASCADE CONSTRAINTS;

CREATE TABLE ResumenEstadisticaCursos AS
SELECT idempleado, nombre, apellidos, salario,COUNT(idcurso) AS total_cursos_impartidos,AVG(num_alumnos) AS promedio_alumnos,MAX(num_alumnos) AS max_alumnos,MIN(num_alumnos) AS min_alumnos,AVG(total_cobrado) AS promedio_cobrado,MAX(total_cobrado) AS max_cobrado,MIN(total_cobrado) AS min_cobrado
FROM InformacionMatriculaCursos
GROUP BY idempleado, nombre, apellidos, salario;

DROP TABLE EstadisticasResumen CASCADE CONSTRAINTS;

CREATE TABLE EstadisticasResumen AS
SELECT idempleado, nombre, apellidos, salario,COUNT(idcurso) AS total_cursos,AVG(alumnos) AS avg_alumnos,MAX(alumnos) AS max_alumnos,MIN(alumnos) AS min_alumnos,AVG(recaudado) AS avg_recaudado,MAX(recaudado) AS max_recaudado,MIN(recaudado) AS min_recaudado
FROM (
    SELECT e.idempleado, e.nombre, e.apellidos, e.salario, c.idcurso,COUNT(cc.idcliente) AS alumnos,(COUNT(cc.idcliente) * c.precio) AS recaudado
    FROM EMPLEADOS e JOIN CURSOS c ON e.idempleado = c.instructor JOIN CURSOS_CLIENTES cc ON c.idcurso = cc.idcurso
    GROUP BY e.idempleado, e.nombre, e.apellidos, e.salario, c.idcurso, c.precio
)
GROUP BY idempleado, nombre, apellidos, salario;

