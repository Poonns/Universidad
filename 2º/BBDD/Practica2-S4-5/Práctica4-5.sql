/*Ejercicio 0: Preparación de las tablas para los ejercicios. El alumno debe crear una copia de las
tablas antes mencionadas, de esta forma todos trabajareis con el mismo modelo. Para ello podemos
utilizar un fichero llamado, por ejemplo, "crear_tablas_plsql.sql" que borre las que hayáis creado en
la práctica 1 y cree las copias nuevas. En el caso de que alguna no haya sido creada anteriormente os
aparecerá un mensaje de error, pues la tabla a borrar no existe.*/

DROP TABLE mi_clientes CASCADE CONSTRAINTS;
DROP TABLE mi_mascotas CASCADE CONSTRAINTS;
DROP TABLE mi_animales CASCADE CONSTRAINTS;
DROP TABLE mi_cursos CASCADE CONSTRAINTS;
DROP TABLE mi_cursos_clientes CASCADE CONSTRAINTS;
DROP TABLE mi_empleados CASCADE CONSTRAINTS;
DROP TABLE mi_pedidos CASCADE CONSTRAINTS;
DROP TABLE mi_pedidos_productos CASCADE CONSTRAINTS;
DROP TABLE mi_productos CASCADE CONSTRAINTS;
DROP TABLE mi_proveedores CASCADE CONSTRAINTS;

CREATE TABLE mi_clientes AS SELECT * FROM clientes;
CREATE TABLE mi_mascotas AS SELECT * FROM mascotas;
CREATE TABLE mi_animales AS SELECT * FROM animales;
CREATE TABLE mi_cursos AS SELECT * FROM cursos;
CREATE TABLE mi_cursos_clientes AS SELECT * FROM cursos_clientes;
CREATE TABLE mi_empleados AS SELECT * FROM empleados;
CREATE TABLE mi_pedidos AS SELECT * FROM pedidos;
CREATE TABLE mi_pedidos_productos AS SELECT * FROM pedidos_productos;
CREATE TABLE mi_productos AS SELECT * FROM productos;
CREATE TABLE mi_proveedores AS SELECT * FROM proveedores;

/*Ejercicio 1: Escribe un bloque que obtenga en una variable la cantidad de productos en stock que
existen en el sistema y la saque por la salida.*/
SET SERVEROUTPUT ON;

DECLARE 
    cantidad number;
BEGIN
    select sum(stock) into cantidad from mi_productos;
    dbms_output.put_line('El stock total es ' || cantidad);
END;

/*Ejercicio 2: Modificar el código del procedimiento PR02 para que muestre la información de forma
tabular: una fila por tabla existente en el esquema que muestre el nombre de la tabla y el número de
registros que contiene. Debe escribir una línea final con el número total de registros que hay entre
todas las tablas del esquema. El procedimiento se llamará pr02a.
Nota: Se trata de añadir líneas de escritura después de cada acceso a las distintas tablas.
Ejemplo de salida:
CLIENTES: 20 filas
MASCOTAS: 10 filas
ANIMALES: 8 filas
…*/

CREATE OR REPLACE PROCEDURE PR02a IS
    xparcial number := 0;
    xtotal number := 0;
BEGIN
    select count(*) into xparcial from mi_clientes;
    xtotal := xtotal + xparcial;
    dbms_output.put_line('Clientes: '|| xparcial || 'filas');
    select count(*) into xparcial from mi_mascotas;
    xtotal := xtotal + xparcial;
    dbms_output.put_line('Macotas: '|| xparcial || 'filas');
    select count(*) into xparcial from mi_animales;
    xtotal := xtotal + xparcial;
    select count(*) into xparcial from mi_cursos;
    xtotal := xtotal + xparcial;
    dbms_output.put_line('Cursos: '|| xparcial || 'filas');
    select count(*) into xparcial from mi_cursos_clientes;
    xtotal := xtotal + xparcial;
    dbms_output.put_line('Cursos_clientes: '|| xparcial || 'filas');
    select count(*) into xparcial from mi_empleados;
    xtotal := xtotal + xparcial;
    dbms_output.put_line('Empleados: '|| xparcial || 'filas');
    select count(*) into xparcial from mi_pedidos;
    xtotal := xtotal + xparcial;
    dbms_output.put_line('Pedidos: '|| xparcial || 'filas');
    select count(*) into xparcial from mi_pedidos_productos;
    xtotal := xtotal + xparcial;
    dbms_output.put_line('Pedidos_productos: '|| xparcial || 'filas');
    select count(*) into xparcial from mi_productos;
    xtotal := xtotal + xparcial;
    dbms_output.put_line('Productos: '|| xparcial || 'filas');
    select count(*) into xparcial from mi_proveedores;
    xtotal := xtotal + xparcial;
    dbms_output.put_line('Proveedores: '|| xparcial || 'filas');

    dbms_output.put_line('En total hay ' || to_char(xtotal) ||' registros');
END PR02a;

/*Ejercicio 3: Realizar un procedimiento que borre todos los proveedores con los que no se está
trabajando actualmente (es decir aquellos que no aparecen en productos). Debes añadir también el
manejo de excepciones.*/
CREATE OR REPLACE PROCEDURE PR03 IS
BEGIN
    delete from proveedores
        where idproveedor not in (select idproveedor from productos);
    COMMIT;
EXCEPTION
    when other then 
    rollback;
END;

/*Ejercicio 4a:
1. ¿Por qué usamos una variable "OUT"? Cambia el tipo de parámetro y prueba el código. ¿Qué
ocurre?
2. ¿Qué ocurre si ejecutamos el procedimiento para un código de curso que no existe? Añade el
tratamiento de excepciones de forma adecuada. ¿Es entonces correcto el código anterior?*/
CREATE OR REPLACE PROCEDURE pr04
(xcodcurs IN mi_cursos_clientes.idcurso%type, xtotal OUT number) IS
BEGIN
    select count(idcliente) into xtotal
FROM mi_cursos_clientes where idcurso = xcodcurs;
    dbms_output.put_line ('Alumnos inscritos: ' || xtotal);
EXCEPTION
    when others then
        xtotal := NULL;
        dbms_output.put_line('Ocurrió un error al contar los alumnos.');
END pr04;

/*Ejercicio 4b. Modificar el código del procedimiento pr04 apropiadamente para transformarlo en una
función equivalente. La función se debe llamar fn04. Escribe las sentencias SQL necesarias para
probar el procedimiento y la función.
Nota: La función tendrá un parámetro de entrada (el código de curso) y devolverá el número
de alumnos inscritos.*/
CREATE OR REPLACE FUNCTION fn04
(xcodcurs IN mi_cursos_clientes.idcurso%type, xtotal OUT number) return number is
BEGIN
    select count(idcliente) into xtotal
    FROM mi_cursos_clientes where idcurso = xcodcurs;
    dbms_output.put_line ('Alumnos inscritos: ' || xtotal);
    return xtotal;
END fn04;

select idcliente, fn04(idcliente) as clintes from mi_cursos_clientes;

/*Ejercicio 5. En este caso vamos a crear una función que nos permita añadir un pedido de un producto
acordado al sistema. La función debe realizar comprobaciones como que el cliente existe, el producto
existe, que hay existencias suficientes del producto, etc. Además, se debe comprobar que el empleado
es correcto, así como que las fechas de entrega sean válidas (posteriores o iguales al día del pedido,
fecha tomada del sistema).
El código siguiente está incompleto. Escribe el código necesario para que realice lo que se pide en el
ejercicio 5. El código faltante no tiene que ver con el uso de las funciones que sirven para identificar
el error (SQLCODE y SQLERRM).
Prueba la función con distintos valores. Por ejemplo, intenta insertar dos veces los mismos datos.
¿Qué ocurre? Cambia el cliente e inserta los mismos datos, pero para un cliente diferente. ¿Qué
ocurre? ¿Es correcto el funcionamiento o debería ser otro?*/
CREATE OR REPLACE FUNCTION fn05 (
   xidcli mi_pedidos.idcliente%type,
   xfechaentrega VARCHAR2,
   xdirentrega mi_pedidos.direccionentrega%type,
   xemp mi_pedidos.empleadoventa%type,
   xproducto mi_productos.idproducto%type,
   xcantidad mi_pedidos_productos.cantidad%type,
   xpvpreal mi_pedidos_productos.precioreal%type
) RETURN NUMBER IS


   wcliente number;
   wempventa number;
   wstock number;
   wproducto number;
  
   -- VARIABLES AÑADIDAS: Necesarias para la lógica que faltaba
   v_idpedido number;
   v_fecha_entrega_date date;


   -- EXCEPCIONES: Declaramos las que faltaban en tu enunciado
   error_cliente exception;
   error_empleado exception;
   error_producto exception;
   error_stock exception;
   error_fecha exception;


BEGIN
   -- 1. Contar si existen los implicados en las tablas
   SELECT COUNT(*) INTO wcliente FROM mi_clientes WHERE idcliente = xidcli;
   SELECT COUNT(*) INTO wempventa FROM mi_empleados WHERE idempleado = xemp;
  
   -- AÑADIDO: Faltaba contar el producto antes de comprobarlo
   SELECT COUNT(*) INTO wproducto FROM mi_productos WHERE idproducto = xproducto;


   -- Lanzamos las excepciones de existencia
   IF wcliente = 0 THEN RAISE error_cliente; END IF;
   IF wempventa = 0 THEN RAISE error_empleado; END IF;
   IF wproducto = 0 THEN RAISE error_producto; END IF;


   -- 2. Comprobación de Stock
   -- (Ojo: esto debe ir DESPUÉS de comprobar que el producto existe para evitar errores)
   SELECT stock INTO wstock FROM mi_productos WHERE idproducto = xproducto;
   IF wstock < xcantidad THEN
       RAISE error_stock;
   END IF;


   -- 3. Comprobación de la fecha de entrega (AÑADIDO)
   -- Convertimos el texto (VARCHAR2) a una fecha real para poder comparar
   v_fecha_entrega_date := TO_DATE(xfechaentrega, 'DD/MM/YYYY');
  
   -- TRUNC(SYSDATE) nos da la fecha de hoy sin las horas y minutos
   IF v_fecha_entrega_date < TRUNC(SYSDATE) THEN
       RAISE error_fecha;
   END IF;

  
   -- Generamos un nuevo ID para el pedido (el máximo actual + 1)
   SELECT NVL(MAX(idpedido), 0) + 1 INTO v_idpedido FROM mi_pedidos;


   -- A) Insertar en la cabecera del pedido
   INSERT INTO mi_pedidos (idpedido, idcliente, fechapedido, fechaentrega, direccionentrega, empleadoventa)
   VALUES (v_idpedido, xidcli, TRUNC(SYSDATE), v_fecha_entrega_date, xdirentrega, xemp);


   -- B) Insertar en las líneas de detalle del pedido
   INSERT INTO mi_pedidos_productos (idpedido, idproducto, cantidad, precioreal)
   VALUES (v_idpedido, xproducto, xcantidad, xpvpreal);


   -- C) Actualizar el almacén (restar el stock)
   UPDATE mi_productos
   SET stock = stock - xcantidad
   WHERE idproducto = xproducto;


   -- Si todo ha ido bien hasta aquí, guardamos permanentemente
   COMMIT;
   DBMS_OUTPUT.PUT_LINE('Pedido grabado correctamente en el sistema');
   RETURN 1;


EXCEPTION
   -- Manejo de errores corregido y ampliado
   WHEN error_cliente THEN
       DBMS_OUTPUT.PUT_LINE('Error - 1: El cliente no existe. Primero debe registrarse en el sistema');
       RETURN -1;
   WHEN error_empleado THEN
       DBMS_OUTPUT.PUT_LINE('Error - 2: El empleado no existe.');
       RETURN -2;
   WHEN error_producto THEN
       DBMS_OUTPUT.PUT_LINE('Error - 3: Producto desconocido o inexistente');
       RETURN -3;
   WHEN error_stock THEN
       DBMS_OUTPUT.PUT_LINE('Error - 4: No hay stock suficiente del producto.');
       RETURN -4;
   WHEN error_fecha THEN
       DBMS_OUTPUT.PUT_LINE('Error - 5: Fecha de entrega anterior a fecha pedido');
       RETURN -5;
   WHEN OTHERS THEN
       DBMS_OUTPUT.PUT_LINE('Error - 10: error no conocido');
       DBMS_OUTPUT.PUT_LINE('Error Oracle: ' || TO_CHAR(SQLCODE) || ' Mensaje: ' || SUBSTR(SQLERRM, 1, 200));
       ROLLBACK;
       RETURN -10;
END fn05;


/*Ejercicio 6: Realiza un procedimiento (pr10) que inserte un nuevo empleado. Muestra que el
procedimiento funciona correctamente para todos los casos posibles que se pudiesen presentar.*/
CREATE OR REPLACE PROCEDURE pr10 (
   x_idempleado IN mi_empleados.idempleado%type,
   x_nombre     IN mi_empleados.nombre%type,
   x_apellidos  IN mi_empleados.apellidos%type,
   x_email      IN mi_empleados.email%type,
   x_direccion  IN mi_empleados.direccion%type,
   x_telefono   IN mi_empleados.telefono%type,
   x_trabajo    IN mi_empleados.trabajo%type,
   x_salario    IN mi_empleados.salario%type
) IS
BEGIN
   -- Intentamos hacer la inserción directamente
   INSERT INTO mi_empleados (idempleado, nombre, apellidos, email, direccion, telefono, trabajo, salario)
   VALUES (x_idempleado, x_nombre, x_apellidos, x_email, x_direccion, x_telefono, x_trabajo, x_salario);
  
   COMMIT;
   DBMS_OUTPUT.PUT_LINE('ÉXITO: Empleado ' || x_nombre || ' insertado correctamente.');


EXCEPTION
   -- Caso posible 1: El ID ya existe (violación de clave primaria)
   WHEN DUP_VAL_ON_INDEX THEN
       DBMS_OUTPUT.PUT_LINE('ERROR: Ya existe un empleado con el ID ' || x_idempleado || '.');
       ROLLBACK;
      
   -- Caso posible 2: Cualquier otro error (ej. texto demasiado largo, letra en campo numérico...)
   WHEN OTHERS THEN
       DBMS_OUTPUT.PUT_LINE('ERROR CRÍTICO: ' || SQLERRM);
       ROLLBACK;
END pr10;

/*Ejercicio 7: Modifica el procedimiento anterior de tal forma que si ya existe el empleado (es decir,
si existe con el mismo nombre), borre sus datos e introduzca de nuevo sus datos modificados.*/
CREATE OR REPLACE PROCEDURE pr10_v2 (
   x_idempleado IN mi_empleados.idempleado%type,
   x_nombre     IN mi_empleados.nombre%type,
   x_apellidos  IN mi_empleados.apellidos%type,
   x_email      IN mi_empleados.email%type,
   x_direccion  IN mi_empleados.direccion%type,
   x_telefono   IN mi_empleados.telefono%type,
   x_trabajo    IN mi_empleados.trabajo%type,
   x_salario    IN mi_empleados.salario%type
) IS
   v_contador NUMBER;
BEGIN
   -- 1. Miramos si ya existe alguien con ese mismo nombre en la base de datos
   SELECT COUNT(*) INTO v_contador
   FROM mi_empleados
   WHERE nombre = x_nombre;
  
   -- 2. Si el contador es mayor que 0, significa que existe. Lo borramos.
   IF v_contador > 0 THEN
       DELETE FROM mi_empleados WHERE nombre = x_nombre;
       DBMS_OUTPUT.PUT_LINE('AVISO: El empleado ' || x_nombre || ' ya existía. Sus datos antiguos han sido borrados.');
   END IF;


   -- 3. Ahora que el terreno está limpio (o si no existía de antes), insertamos el nuevo
   INSERT INTO mi_empleados (idempleado, nombre, apellidos, email, direccion, telefono, trabajo, salario)
   VALUES (x_idempleado, x_nombre, x_apellidos, x_email, x_direccion, x_telefono, x_trabajo, x_salario);
  
   COMMIT;
   DBMS_OUTPUT.PUT_LINE('ÉXITO: Datos de ' || x_nombre || ' guardados en el sistema.');


EXCEPTION
   WHEN OTHERS THEN
       DBMS_OUTPUT.PUT_LINE('ERROR CRÍTICO: ' || SQLERRM);
       ROLLBACK;
END pr10_v2;

/*Ejercicio 8: Copia el siguiente bloque PL/SQL. Modifícalo, de forma que trabaje adecuadamente
con cursores (el cursor se debe abrir, recuperar datos y cerrar). Si hay algún error adicional, corrígelo
también. ¿Qué es lo que realiza este procedimiento? ¿Se podría haber hecho con una sentencia SQL?
Razona tu respuesta.*/

--bloque pl/sql con uso de cursores
declare
cursor c1 is
select tipoanimal, count(*), avg(precio)
from mi_cursos
group by tipoanimal;

xtotal_cursos number;
xtipoanimal mi_cursos.tipoanimal%type;
xcuenta_curso_animal number;
xprecio_medio_curso_animal number;
total_porcentaje number;

BEGIN
select count(*) into xtotal_cursos from mi_cursos;
open c1;
loop
fetch c1 into xtipoanimal, xcuenta_curso_animal, xprecio_medio_curso_animal;
exit when c1%NOTFOUND;
--se trata los datos recuperados
total_porcentaje:= xcuenta_curso_animal/xtotal_cursos*100;
dbms_output.put_line(rpad(xtipoanimal,10,' ')
||lpad(to_char(xcuenta_curso_animal,'999999'),10,' ') ||
lpad(to_char(total_porcentaje,'99999.99'),10,' ') ||
lpad(to_char(xprecio_medio_curso_animal,'999999'),10,' '));
end loop;
close c1;

EXCEPTION
WHEN OTHERS THEN
dbms_output.put_line('Error -10: error no conocido');
dbms_output.put_line('Error Oracle ' || TO_CHAR(SQLCODE) ||
'Mensaje: ' || SUBSTR(SQLERRM, 1, 200));
END;

-- El procedimiento muestra el nombre del tipo animal, cuantos cursos hay de ese animal y que porcentaj representan esos cursos en el total de cursos que hay en la academia
-- Si que se podria implementar con una sentencia sql

/*Ejercicio 9: Realiza un procedimiento (pr07) que inserte un nuevo curso y que lo imparta el instructor
con menor número de cursos, en el caso de igual número se le asignará al más joven. Muestra que el
procedimiento funciona correctamente para todos los casos posibles que se pudiesen presentar.
Modifica el procedimiento de tal forma que si ya existe el curso (es decir, si existe con el mismo
nombre), borre sus datos e introduzca de nuevo sus datos modificados. NOTA: Se necesita un campo
fechanac en la tabla EMPLEADOS. Es necesario añadir este atributo y actualizar sus valores.*/

alter table mi_empleados
add(fechanac DATE);

update mi_empleados
set fechanac = to_date('10/02/2006', 'DD/MM/YYYY' )
where idempleado = 1;

create or replace procedure pr07 (
    p_idcurso     mi_curso.idcurso%type,
    p_titulo      mi_curso.titulo%type,
    p_descripcion mi_curso.descripción%type,
    p_max_num     mi_curso.max_num%type,
    p_precio      mi_curso.precio%type,
    p_tipoanimal  mi_curso.tipoanimal%type,
    p_fechahora   mi_curso.fechahora%type,
    p_lugar       mi_curso.lugar%type
) IS
    v_instructor_elegido mi_empleado.idempleado%type;
begin
    select instructor into v_instructor_elegido
    from (select instructor from curso c join empleado e on c.instructor = e.idempleado
          group by c.instructor, e.fechanac order by count(c.idcurso) ASC, e.fechanac DESC)
    where rownum = 1;

    insert into mi_curso(idcurso, titulo, descripcion, max_num, precio, tipoanimal, fechahora, lugar, instructor)
    values (p_idcurso, p_titulo, p_descripcion, p_max_num, p_precio, p_tipoanimal, p_fechahora, p_lugar, v_instructor_elegido)

    commit;
    dbms_output.put_line('Curso insertado correctamente con el instructor: ' || v_instructor_elegido);

exception
    when others then
        rollback;
        dbms_output.put_line('Error al insertar el curso ' || SQLERRM);
end pr07;

/*Ejercicio 10: Escribe una función que calcule el precio promedio de los pedidos pendientes de servir
para cada cliente. Utiliza dicha función después en una consulta SELECT que muestre para cada
cliente aquellos pedidos cuyo precio sea mayor al promedio (se mostrará también la información del
cliente que lo ha realizado*/
create or replace function fn10(p_idcliente pedidos.idcliente%type) return number IS
    promedio number;
begin
    select avg(pp.precioreal) into promedio from pedidos_productos pp 
    join pedidos p on pp.idpedido = p.idpedido join clientes c on p.idcliente = c.idcliente
    where c.idcliente = p_idcliente and fechaentrega > sysdate
    return promedio;
end fn10;

select fn10(c.idcliente) as precio_medio_cliente, c.idcliente from cliente c join pedidios p on c.idcliente = p.idcliente
join productos_pedidos pp on p.idpedido = pp.idpedido
where pp.precioreal > fn10(c.idcliente);

/*Ejercicio 11: Realizar una función/procedimiento PL/SQL que permita modificar un pedido
pendiente de servir, al que se le desea añadir un nuevo producto (a ese pedido). Para ello, se le pasará
como argumento el nombre y apellidos del cliente, la fecha del pedido a modificar, así como la
descripción del producto a añadir. Los atributos necesarios para completar la inserción se deben
obtener a partir de los argumentos pasados al procedimiento.*/

create or replace procedure pr11(
    p_nombre        clientes.nombre%type,
    p_apellidos     clientes.apellidos%type,
    p_fechapedido   pedidos.fechapedido%type,
    p_descripcion   pedidos.descripcion%type
) IS
    v_idcliente     clientes.idcliente%type;
    v_idproducto    productos.idproducto%type;
    v_idpedido      pedidos.idpedido%type;
    v_precioventa   productos.precio_venta%type;
begin 
    select idcliente into v_idcliente from clientes
    where nombre = p_nombre and apellidos = p_apellidos

    select idpedido into v_idpedido from pedidos 
    where idcliente = v_idcliente and fechapedido = p_fechapedido and fechaentrega > sysdate

    select idproducto, precio_venta into v_idproducto, v_precioventa
    from productos
    where descripcion = p_descripcion

    insert into pedidos_productos(idpedido, idproducto, cantidad, precioreal)
    values(v_idpedido, v_idproducto, 1, v_precioventa)

    commit
    dbms_output.put_line('Producto introducido correctamente');
exception
    when others then
    rollback;
    dbms_output.put_line('Error al insertar el producto' || SQLERRM);
end;

/*Ejercicio 12: La empresa va a realizar una campaña de promoción para fomentar las ventas. Para
ello, a todos los clientes con compras superiores a 50€ (en pedidos realizados en el último mes) les
aplicará un descuento de un 10% en aquellos pedidos pendientes de servir. Además, también se desea
bonificar a aquellos clientes que estén matriculados en más de 1 curso con un descuento de un 15 %
sobre el precio del curso más económico, para ello habrá que al añadir un nuevo atributo a la tabla
CURSOS_CLIENTES y actualizarlo con el precio real que debería pagar. Realizar una función o
procedimiento que lleve a cabo estas acciones.*/
alter table cursos_clientes
add(precio_total NUMBER)

create or replace procedure pr12 IS
    cursor c1 IS select idcliente 
    from cursos_cliente
    group by idcliente
    having count(*) > 1

    v_idcurso_mas_barato cursos.precio%type
begin
    update pedidos_productos
    set precio real precioreal * 0.9
    where idpedido in (select idpedido from pedidos
    where fechaentrega is null and idcliente in (select p.idcliente 
    from pedidos_productos pp join pedidos p on pp.idpedido = p.pedido
    where fechaentrega > sysdate and fechapedido >= sysdate - 30
    group by pp.idepedido
    having sum(pp.cantidad * pp.precioreal) > 50))

    FOR reg IN c1 LOOP
        SELECT idcurso INTO v_idcurso_mas_barato
        FROM (
            SELECT cc.idcurso
            FROM cursos_clientes cc
            JOIN cursos c ON cc.idcurso = c.idcurso
            WHERE cc.idcliente = reg.idcliente
            ORDER BY c.precio ASC
        )
        WHERE rownum = 1;

        UPDATE cursos_clientes
        SET precio_total = (SELECT precio * 0.85 FROM cursos WHERE idcurso = v_idcurso_mas_barato)
        WHERE idcliente = reg.idcliente 
          AND idcurso = v_idcurso_mas_barato;
    END LOOP;

/*Ejercicio 13: Realiza un procedimiento/función PL/SQL al que se le pase como parámetro el nombre
de un instructor y muestre información de cada uno de los cursos impartidos, indicando el precio y
que clientes están inscritos agrupados por su ciudad.*/
create or replace procedure pr13(p_instructor cursos.instructor%type) IS
    cursor c1 is select idcurso, titulo, precio into v_idcurso, v_titulo, v_precio
    from cursos
    where instructor = p_instructor

    v_precio cursos.precio%type;
    v_idcurso cursos.idcurso%type;
    v_titulo cursos.titulo%type;
    v_idcliente cursos_clientes.idcliente%type;
    v_lugar cursos.lugar%type;
begin
    for reg in c1 loop
        dbms_output.put_line('Curso: '|| reg.idcurso || 'Titulo: '|| reg.titulo || 'Precio: '|| reg.precio);
        for reg1 in (select idcliente, lugar 
            from cursos c join cursos_cliente cc on c.idcliente = cc.idcliente
            where cc.idcurso = reg.idcurso
            order by c.lugar) loop
            dbms_outpunt.put_line('Cliente: '|| idcliente || 'Ciudad: '|| v_lugar);
        end loop;
    end loop;   

exception
    when others THEN
    dbms_output.put_line('Error. '||SQLERRM)
end;





--Extra 1
create or replace procedure pr_reposicion_automatica is
    cursor c1 is select idproducto, stockactual, stockmin, precio
    from productos

    v_contador number := 0;
    v_cantidad number;
begin
    for reg in c1 loop
        if reg.stockactual < reg.stockmin then
            if reg.precio > 100 then
                v_cantidad := 50;
            else
                v_cantidad := 100;
            end if;

            insert into pedidos_proveedores( id_prod, cantidad_pedida, fecha)
            values(reg.id_prod, v_cantidad, sysdate)
            v_contador := v_contador + 1;
        end if;
    end loop;

    dbms_output.put_line('Pedidos totales: '|| v_contador);
exception
    when others then 
    dbms_output.put_line('Error. '|| SQLERRM);
end;

--Extra 2
create or replace procedure prod(p_ciudad departamentos.ciudad%type) IS
    cursor c1 is select nombre_dep, id_dep
    from departamentos
    where ciudad = p_ciudad

    v_total number := 0;
begin
    for reg in c1 loop
        for reg1 in (select e.id_emp, e.salario from empleados e join departamentos d on e.id_dep = d.id_dep where e.id_dep = reg.dep) loop
        dbms_output.put_line('Empleado: ' || reg1.id_emp || ' Salario: '||reg1.salario);
        v_total := v_total + reg1.salario;
        end loop;
        dbms_output.put_line('Total Salarios: ' || v_total||);
        v_total := 0;
    end loop;
exception
    when others then 
        dbms_output.put_line('Error. ' || SQLERRM);
end;

/*Ejercicio 14: Realiza un procedimiento/función PL/SQL al que se le pase como parámetro el nombre
de un cliente y muestre información de cada uno de los pedidos realizados, indicando que productos
incluye ese pedido, cuánto costó cada producto del pedido, la cantidad de ese producto en el pedido
y el precio total. Se le puede pasar también un periodo de tiempo, de forma que muestre solo esos
pedidos. Se muestra un ejemplo de ejecución:*/
CREATE OR REPLACE PROCEDURE pr14(
    p_nom_cliente IN clientes.nombre%TYPE,
    p_fechaIni    IN DATE DEFAULT SYSDATE - 60, -- Sintaxis correcta
    p_fechaFin    IN DATE DEFAULT SYSDATE
) IS
    -- Cursor para encontrar al cliente (por si hay varios con el mismo nombre)
    CURSOR c_clientes IS 
        SELECT idcliente FROM clientes WHERE nombre = p_nom_cliente;
    
    v_total_pedido NUMBER;
BEGIN
    FOR r_cli IN c_clientes LOOP
        -- NIVEL 1: Info del Cliente
        dbms_output.put_line('Pedidos de: ' || p_nom_cliente);

        -- NIVEL 2: Buscamos sus pedidos en las fechas dadas
        FOR r_ped IN (
            SELECT idpedido, fechapedido 
            FROM pedidos 
            WHERE idcliente = r_cli.idcliente 
              AND fechapedido BETWEEN p_fechaIni AND p_fechaFin
        ) LOOP
            dbms_output.put_line('  Pedido: ' || r_ped.idpedido || ' Fecha: ' || r_ped.fechapedido);
            v_total_pedido := 0; -- Reiniciamos el total para cada pedido nuevo

            -- NIVEL 3: Los productos de ESE pedido
            FOR r_prod IN (
                SELECT pp.idproducto, pr.nombre_prod, pp.precioreal, pp.cantidad
                FROM pedidos_productos pp
                JOIN productos pr ON pp.idproducto = pr.idproducto
                WHERE pp.idpedido = r_ped.idpedido
            ) LOOP
                -- Calculamos subtotal de la línea
                v_total_pedido := v_total_pedido + (r_prod.precioreal * r_prod.cantidad);
                
                dbms_output.put_line('    - ' || r_prod.nombre_prod || ' | Cant: ' || r_prod.cantidad || ' | Precio: ' || r_prod.precioreal);
            END LOOP;

            dbms_output.put_line('  TOTAL DEL PEDIDO: ' || v_total_pedido || '€');
            dbms_output.put_line('---------------------------------------');
        END LOOP;
    END LOOP;
END;