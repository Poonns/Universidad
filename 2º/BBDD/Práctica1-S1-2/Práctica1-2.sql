/*Ejercicio 16: Obtener los pedidos con fecha de entrega en el 2017 (es decir comprendidos
entre el 1/01/2017 y el 31/12/2017).*/
SELECT idpedido
FROM PEDIDOS
WHERE EXTRACT(YEAR FROM fechaentrega) = 2017;

/*Ejercicio 17: Obtener el salario mensual neto de los empleados que han impartido cursos
sobre gatos, suponiendo un IRPF del 20% (el titulo debe contener la palabra gato, tanto en
mayúsculas como minúsculas).*/
SELECT e.nombre, e.apellidos, e.salario * 0.8 / 12 AS salario_neto
FROM EMPLEADOS e JOIN CURSOS cr ON e.idempleado = cr.instructor
WHERE UPPER(cr.titulo) LIKE '%GATO%';

/*Ejercicio 18: Indicar el precio mínimo, máximo y medio de los cursos según el tipo de animal.*/
SELECT an.nombre, MIN(cr.precio), MAX(cr.precio), AVG(cr.precio)
FROM CURSOS cr JOIN ANIMALES an ON cr.tipoanimal = an.tipoanimal
GROUP BY an.nombre;

/*Ejercicio 19: Obtén los datos de los clientes que sean de Burjassot y muéstralos ordenados
alfabéticamente por apellidos y nombre.*/
SELECT *
FROM CLIENTES 
WHERE ciudad = 'Burjassot'
ORDER BY apellidos ASC, nombre ASC;

/*Ejercicio 20: Escribe una consulta que obtenga, para cada tipo de trabajo, el número de
empleados que cobra más que el salario medio de todos los empleados.*/
SELECT e.trabajo, COUNT(e.idempleado)
FROM EMPLEADOS e
WHERE e.salario > (SELECT AVG(salario)
    FROM EMPLEADOS)
GROUP BY e.trabajo;

/*Ejercicio 21: Encuentra el número de cursos que se han ofertado para cada mes del año 2017
Es decir se desea saber cuántos cursos se han ofertado en ENERO, cuántos en FEBRERO, etc.*/
SELECT EXTRACT(MONTH FROM fechahora) AS mes, COUNT(idcurso)
FROM CURSOS cr
WHERE EXTRACT(YEAR FROM fechahora) = 2017
GROUP BY EXTRACT(MONTH FROM fechahora)
ORDER BY mes;

/*Ejercicio 22: Obtener el nombre de los proveedores que no suministran ningún producto.*/
SELECT pr.nombre
FROM PROVEEDORES pr LEFT OUTER JOIN PRODUCTOS prod ON pr.ideproveedor = prod.ideproveedor
WHERE prod.idproducto IS NULL;

/*Ejercicio 23: Encuentra el nombre de los empleados que han impartido más de 3 cursos.*/
SELECT e.nombre, e.apellidos, COUNT(idcurso) AS total_cursos
FROM EMPLEADOS e JOIN CURSOS cr ON e.idempleado = cr.instructor
GROUP BY e.ideempleado, e.nombre, e.apellidos
HAVING COUNT(idcurso) > 3;

/*Ejercicio 24: Obtener el nombre y el salario de los empleados que cobran más que el
empleado con nombre ‘Jose’ y menos que el empleádo cuyo nombre es ‘Clara’.*/
SELECT e.nombre, e.salario
FROM EMPLEADOS e
WHERE e.salario > (SELECT e1.salario
    FROM EMPLEADOS e1
    WHERE e1.nombre = 'Jose') AND e.salario < (SELECT e2.salario 
    FROM EMPLEADOS e2
    WHERE e2.nombre = 'Clara');

/*Ejercicio 25: Obtener el nombre y el salario de los empleados que cobran más que el
empleádo con nombre ‘Jose’ y menos que los empleádos cuyo emáil es de ‘gmáil’.*/
SELECT e.nombre, e.salario
FROM EMPLEADOS e 
WHERE e.salario > (SELECT e.salario
    FROM EMPLEADOS e
    WHERE e.nombre = 'Jose') AND e.salario < ALL (SELECT e.salario
    FROM EMPLEADOS e
    WHERE e.email LIKE '%gmail%');

/*Ejercicio 26: Obtener el nombre y el salario del empleado que más cobra.*/
SELECT e.nombre, e.salario
FROM EMPLEADOS e
WHERE e.salario = (SELECT MAX(salario)
    FROM EMPLEADOS);

/*Ejercicio 27: Obtén el nombre y el salario de aquél empleado que cobra más en cada uno de
los tipos de trabajos diferentes.*/
SELECT e.nombre, e.salario
FROM EMPLEADOS e
WHERE e.salario = (SELECT MAX(salario)
    FROM EMPLEADOS e2
    WHERE e.trabajo = e2.trabajo);

/*Ejercicio 28: Escribe una consulta SQL para obtener los clientes (nombre y dirección) que
tienen todavía pedidos sin servir y fueron pedidos hace más de quince días (pedidos sin servir
son aquellos que no tienen fecha de entrega).*/
SELECT c.nombre, c.dirección
FROM CLIENTES c JOIN PEDIDOS p ON c.idcliente = p.idcliente
WHERE p.fechaentrega IS NULL AND (SYSDATE - p.fechapedido) > 15;

/*Ejercicio 29: Mostrar aquellos productos pedidos en mayo de 2017 que han subido de precio
(mostrar el idproducto, su descripción y el precio de venta actual).*/
SELECT prod.idproducto, prod.descripción, prod.precio_venta
FROM (PRODUCTOS prod JOIN PEDIDOS_PRODUCTOS pp ON prod.idproducto = pp.idproducto) JOIN PEDIDOS p ON pp.idpedido = p.idpedido
WHERE EXTRACT(MONTH FROM p.fechapedido) = 5 AND EXTRACT(YEAR FROM p.fechapedido) = 2017 AND prod.precio_venta > pp.precioreal;

/*Ejercicio 30: Obtener el nombre del cliente que ha realizado el pedido más caro del año 2017*/
SELECT c.nombre
FROM (CLIENTES c JOIN PEDIDOS p ON c.idcliente = p.idcliente) JOIN PEDIDOS_PRODUCTOS pp ON p.idpedido = pp.idpedido
WHERE EXTRACT(YEAR FROM p.fechapedido) = 2017
GROUP BY c.nombre, p.idpedido
HAVING SUM(pp.cantidad * pp.precioreal) >= ALL (SELECT SUM(pp.cantidad * pp.precioreal)
    FROM PEDIDOS_PRODUCTOS pp2 JOIN PEDIDOS p2 ON p2.idpedido = pp2.idpedido
    WHERE EXTRACT(YEAR FROM p2.fechapedido) = 2017
    GROUP BY p2.idepedido);

/*Ejercicio 31: Escribe una consulta SQL para obtener los pedidos que contienen comida para
gatos y comida para periquitos, pero no contiene Amoxicilina (se debe revisar el atributo
descripción de la tabla productos)*/
SELECT p.idpedido
FROM PEDIDOS p JOIN PEDIDOS_PRODUCTOS pp ON p.idepedido = pp.idpedido JOIN PRODUCTOS prod ON pp.idproducto = prod.idproducto
WHERE p.idepedido IN (SELECT pp.idpedido
    FROM PEDIDOS_PRODUCTOS pp JOIN PRODUCTOS prod ON pp.ideproducto = prod.idproducto
    WHERE UPPER(prod.descripcion) LIKE '%COMIDA PARA GATOS%') 
    AND p.idpedido IN 
    (SELECT pp2.idpedido
    FROM PREDIDOS_PRODUCTOS pp2 JOIN PRODUCTOS prod2 ON pp2.idproducto = prod2.idproducto
    WHERE UPPER(prod2.descripcion) LIKE '%COMIDA PARA PERQUITOS%')
    AND p.idpedido NOT IN 
    (SELECT pp3.idpedido
    FROM PREDIDOS_PRODUCTOS pp3 JOIN PRODUCTOS prod3 ON pp3.idproducto = prod3.idproducto
    WHERE UPPER(prod3.descripcion) LIKE '%AMOXICILINA');

/*Ejercicio 32: Mostrar todos los cursos que están por impartirse, y ya no tienen plazas libres a
día de hoy.*/
SELECT c.idcurso
FROM CURSOS c
WHERE c.fechahora > SYSDATE  AND c.max_num = (SELECT COUNT(*)
    FROM CURSOS_CLIENTES cc
    WHERE cc.idcurso = c.idcurso);

/*Ejercicio 33: Muestra los clientes que todavía no han realizado ningún pedido.*/
SELECT *
FROM CLIENTES c JOIN PEDIDOS_CLIENTES pc 
WHERE 