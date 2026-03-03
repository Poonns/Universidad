/*13. Obtener la media d e los sueldos d e los empleados, agrupados por categoría y
departamento.*/
SELECT e.n_cat, e.n_dto, AVG(sueldo)
FROM EMP e 
GROUP BY e.n_cat, e.n_dto;

/*14. Obtener l a categoría que tiene mayor número d e empleados.*/
