/* Ejercicio 1: Escribe una consulta SQL para visualizar toda la información existente en la tabla
CLIENTES */
SELECT * 
FROM CLIENTES;

/* Ejercicio 2: Escribe una consulta SQL que extraiga toda la información de los CLIENTES que
viven en la ciudad de ‘Burjassot’*/
SELECT *
FROM CLIENTES 
WHERE ciudad = 'Burjassot';

/* Ejercicio 3: Escribe una consulta SQL para visualizar todos los productos cuyo precio_
venta es mayor que 20 y se tienen menos de 10 productos en stock*/
SELECT idproducto
FROM PRODUCTOS
WHERE precio_venta < 20 AND stock > 10;

/* Ejercicio 4: Escribe una consulta SQL que permita mostrar el nombre, el apellido, el email de todos
los clientes que se apellidan ‘García’, llamándole al email “dirección de correo”*/
SELECT C.nombre, C.apellidos, C.email AS "dirección de correo"
FROM CLIENTES C
WHERE C.apellidos = 'García';

/*Ejercicio 5: Encuentra, para cada uno de los productos existentes, el margen de beneficio que
tiene la empresa, es decir la diferencia entre el precio de compra al proveedor y el precio de
venta al público*/
SELECT idproducto, precio_compra - precio_venta AS beneficio
FROM PRODUCTOS;

/*Ejercicio 6: Busca los clientes que no tienen email.*/
SELECT C.nombre, C.apellidos 
FROM CLIENTES C
WHERE C.email IS NULL;

/*Ejercicio 7: Encontrar todos los CLIENTES cuyo email es de la universidad de Valencia.*/
SELECT idcliente
FROM CLIENTES
WHERE email LIKE '%alumni.uv.es';

/*Ejercicio 8a: Escribe una consulta SQL que muestre el nombre de las mascotas y su edad
ordenados por fechanac. Se desea mostrar únicamente las mascotas que son ‘Perros’ o ‘Gátos’,
sabiendo que estos animales corresponden a los códigos de tipoanimal 1 y 2.*/
SELECT nombremascota, TIMESTAMPDIFF(YEAR, fechanac, CURDATE()) AS edad --Sacar la edad
FROM MASCOTAS
WHERE tipoanimal IN(1,2)
ORDER BY fechanac DESC;

/*Ejercicio 8b: Escribe una consulta SQL que muestre el identificador y nombre de los cursos
que se imparten en los meses pares de cualquier año.*/
SELECT idcurso, titulo
FROM CURSOS
WHERE MOD(TO_NUMBER(TO_CHAR(fechahora, 'MM')), 2) = 0; -- MOD() = %

/*Ejercicio 9a: Escribe una consulta que devuelva toda la información de los cursos en los que
se han matriculado cada uno de los clientes (muestra únicamente el atributo idcliente).*/
SELECT c.idcliente, descripcion
FROM (
    SELECT *
    FROM CLIENTES c JOIN CURSOS_CLIENTES cursos_cl ON c.idcliente = cursos_cl.idcliente 
    ) c_l
JOIN CURSOS cursos ON c_l.idcurso = cursos.idcurso;

/*Ejercicio 9b: Escribe una consulta que devuelva toda la información de los cursos junto con el
nombre del animal para el que va dirigido, incluyendo los cursos que no van dirigidos a
ningún animal específico.*/
SELECT c.*, a.nombre
FROM CURSOS c LEFT OUTER JOIN ANIMALES a ON (c.tipoanimal = a.tipoanimal);

/*Ejercicio 9c: Repetir la consulta anterior incluyendo los nombres de los animales para los que
no existen cursos.*/
SELECT c.*, a.nombre
FROM CURSOS c RIGHT OUTER JOIN ANIMALES a ON (c.tipoanimal = a.tipoanimal);

/*Ejercicio 10: Escribe una consulta SQL que muestre los pedidos que incluyen los productos
de comida para gatos y de comida para periquitos.*/
SELECT idpedido
FROM PEDIDOS_PRODUCTOS pp JOIN PRODUCTOS pr ON pp.idproducto = pr.idproducto
WHERE pr.descripcion LIKE '%gato%'
INTERSECT 
SELECT idpedido
FROM PEDIDOS_PRODUCTOS pp JOIN PRODUCTOS pr ON pp.idproducto = pr.idproducto
WHERE pr.descripcion LIKE '%periquito%';

/*Ejercicio 11: Encontrar los instructores que no han impartido el curso con código de curso
idcurso = 1.*/
SELECT instructor
FROM CURSOS cr
WHERE instructor NOT IN
    (SELECT instructor
    FROM CURSOS 
    WHERE idcurso = 1);

/*Ejercicio 12: Encontrar la cantidad de cursos que se imparten durante el mes de agosto de
2017.*/
SELECT COUNT(idcurso)
FROM CURSOS 
WHERE EXTRACT(MONTH FROM fechahora) = 8 AND EXTRACT(YEAR FROM fechahora) = 2017;

/*Ejercicio 13: Encuentra el número de cursos que se han ofertado para cada mes del año 2017.
Es decir, se desea saber cuántos cursos se han ofertado en ENERO, cuántos en FEBRERO, etc.*/
SELECT EXTRACT(MONTH FROM fechahora) AS mes, COUNT(idcurso) AS cantidad_cursos
FROM CURSOS
WHERE EXTRACT(YEAR FROM fechahora) = 2017
GROUP BY EXTRACT(MONTH FROM fechahora)
ORDER BY mes;

/*Ejercicio 14: Encuentra el nombre de los empleados que han impartido más de 3 cursos.*/
SELECT e.nombre, e.apellidos, COUNT(cr.idcurso) AS total_cursos
FROM EMPLEADOS e JOIN CURSOS cr ON e.idempleado = cr.instructor
GROUP BY e.idempleado, e.nombre, e.apllidos
HAVING COUNT(cr.idcurso) > 3;

/*Ejercicio 15: Obtener los empleados que su salario es superior al salario medio de todos los
empleados que imparten o han impartidos cursos.*/
SELECT idempleado
FROM EMPLEADOS e
WHERE salario > (SELECT AVG(salario)
    FROM EMPLEADOS e2 JOIN CURSOS cr ON e.ideempleado = cr.instructor)
    WHERE idempleado IN (SELECT instructor
        FROM CURSOS);