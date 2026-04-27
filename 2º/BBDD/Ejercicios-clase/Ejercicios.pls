/*Definir un cursor C1 para seleccionar aquellos empleados que tenían un sueldo
superior a 25.000 € y han sido contratados con anterioridad al 1 de enero de 2010.*/
DECLARE 
    maximo NUMBER(9,3);
    minimo NUMBER(9,3);
    contador NUMBER(9,0);
    sueldo NUMBER(9,3);
    fecha_r CONSTANT DATE := TO_DATE('1/1/2010');

    CURSOR c1 IS
        SELECT n_emp FROM EMP WHERE sueldo > 25.000 AND fecha_alta < fecha_r;
BEGIN
END;

/*Definir un cursor C2 para modificar el sueldo base de la categoría con un porcentaje
determinado para las categorías de los empleados que pertenezcan a un departamento,
y otro cursor C3 para seleccionar un departamento para el cual la media de los sueldos
de los empleados se encuentre entre 2 valores de referencia sueldo_min y
sueldo_max.*/
DECLARE 
    maximo NUMBER(9,3);
    minimo NUMBER(9,3);
    contador NUMBER(9,0);
    sueldo NUMBER(9,3);
    fecha_r CONSTANT DATE := TO_DATE('1/1/2010');

    CURSOR c2 IS
        SELECT n_dto FROM EMP 
        GROUP BY n_dto
        HAVING AVG(sueldo) BETWEEN maximo AND minimo;

    CURSOR c3 (dto NUMBER) IS
        SELECT n_cat, sueldo_base FROM CAT
        WHERE n_cat IN (SELECT n_cat FROM EMP WHERE n_dto = p_dto)
        FOR UPDATE OF sueldo_baso;
BEGIN
END;

/*Crear el código necesario para implementar la siguiente lógica: encontrar aquellos
empleados que tenían un sueldo superior a 25.000 € y han sido contratados con
anterioridad al 1 de enero de 2010, y obtener el  mejor sueldo y el 
peor sueldo. A continuación se debe localizar los departamentos para los que los
sueldos medios se encuentren entre los valores de referencia encontrados. Para estos
departamentos, se modificará el sueldo base de todas las categorías de los empleados
que pertenecen a ellos, incrementándolo en un 10%.*/
maximo := 0;
minimo := 999999.999;

FOR i IN c1 LOOP
    IF i.sueldo > maximo THEN
        maximo := i.sueldo;
    END IF;
    IF i.sueldo < minimo THEN
        minimo := i.sueldo;
    END IF;
END LOOP;

FOR reg_2 IN c2 (maximo, minimo) LOOP
    FOR reg_3 IN c3(reg_2.n_dto) LOOP
        UPDATE cat
        SET sueldo_base = sueldo_base * incremento / 100
        WHERE current of c3; --para hacer la actualización mientras recorres
    END LOOP;
END LOOP;


/*Consulta*/
UPDATE CAT
SET sueldo_base = sueldo_base * 1.1
WHERE n_cat IN (SELECT n_cat FROM emp
    WHERE n_dto IN (SELECT n_dto FROM emp e2
        WHERE sueldo_base > 25000 AND fecha_alta < '01/01/2010' 
            GROUP BY n_dto
            HAVING AVG(sueldo) BETWEEN 
            (SELECT MIN(sueldo) FROM emp WHERE sueldo_base > 25000 AND fecha_alta < '01/01/2010')
            AND
            (SELECT MIN(sueldo) FROM emp WHERE sueldo_base > 25000 AND fecha_alta < '01/01/2010')));
            

