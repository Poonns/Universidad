/*1. Obtener los detalles completos de todos los proveedores de París.*/
SELECT *
FROM proveedores 
WHERE CIUDAD = 'París';

/*2. Obtener el nombre de todos los proyectos que se desarrollen en Barcelona, París,
Amsterdam, Bruselas, Valencia, Madrid, Berlín y Moscú y que tengan en el nombre al
menos dos letras A.*/
SELECT JNOMBRE
FROM proyectos 
WHERE CIUDAD IN ('Barcelona','París','Amsterdam','Bruselas','Valencia','Madrid','Berlín','Moscú') AND JNOMBRE LIKE '%A%A%';

/*3. Obtener todas las tripletas número de proveedor/número de artículo/número de
proyecto tales que el proveedor, el artículo y el proyecto indicados estén todos en la
misma ciudad.*/
SELECT P#, A#, O#
FROM (Obras o JOIN proveedores p ON  o.CIUDAD = p.CIUDAD) JOIN artículos a ON o.CIUDAD = a.CIUDAD;

/*4. Obtener todas las tripletas de nombre de proveedor, artículo y proyecto de tal modo
que el artículo esté siendo enviado por el proveedor a dicho proyecto.*/
SELECT p.SNOMBRE, a.PNOMBRE, o.JNOMBRE
FROM proveedores p JOIN envios e ON p.P# = e.P# JOIN artículos a ON a.A# = e.A# JOIN obras o ON o.O# = e.O#;

/*5. Obtener el nombre y color de los artículos que tengan un peso conocido (no nulo).*/
SELECT a.PNOMBRE, a.COLOR
FROM artículos a 
WHERE a.PESO IS NOT NULL;

/*6. Obtener los números de los proveedores que suministren por lo menos uno de los
artículos suministrados por al menos un proveedor que suministre un artículo cuyo
peso esté entre 100 y 500 gramos, ambos inclusive (suponer que el campo peso de la
tabla artículos se expresa en kilogramos).*/
SELECT P#
FROM envios e
WHERE A# IN (SELECT A#
    FROM envios e 
    WHERE P# IN (SELECT P#
        FROM envios e 
        WHERE A# IN (SELECT A#
            FROM artículos a
            WHERE a.PESO >= 0.1 AND a.PESO <= 0.5)));

/*7. Obtener la cantidad máxima, mínima y promedio de artículos suministrada por cada
proveedor en cada proyecto, indicando el nombre del proyecto y del proveedor,
ordenados por nombre de proveedor y nombre de proyecto.*/
SELECT o.JNOMBRE,p.SNOMBRE, MAX(e.CANTIDAD), MIN(e.CANTIDAD), AVG(e.CANTIDAD)
FROM (envios e JOIN proveedores p ON p.P# = e.P#)JOIN obras o ON o.O# = e.O#
GROUP BY o.JNOMBRE, p.SNOMBRE
ORDER BY p.SNOMBRE, o.JNOMBRE;

/*8. Obtener los números de los proyectos a los cuales se suministre el artículo A1 en una
cantidad promedio mayor que la cantidad máxima en la cual se suministra algún
artículo al proyecto O1.*/
SELECT o.O# 
FROM envios e
WHERE e.A# = 'A1'
GROUP BY o.O#
HAVING AVG(e.CANTIDAD) > (SELECT MAX(e.CANTIDAD)
    FROM envios e
    WHERE e.O# = '01');