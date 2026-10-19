--p1
/**Construir una consulta en SQL para obtener todos los datos de los clientes 
que tengan dirección de email de gmail o no tengan dirección de email y que vivan en la ciudad de Valencia.*/
select * from clientes
where (email like '%gmail.com%' or email is null) and ciudad = 'Valencia';

/*Construir una consulta en SQL para obtener la dirección de entrega en mayúsculas de los pedidos hechos en 
abril de 2024 junto con la dirección postal de los clientes que viven en Burjassot, ordenado alfabéticamente.*/
select UPPER(p.direccionentrega) as direccion_entrega, cli.direccionpostal
from pedidos p join clientes cli on p.idcliente = cli.idcliente
where to_char(p.fechapedido, 'MM/YYYY') = '04/2024' and cli.ciudad = 'Burjassot'
order by direccion_entrega asc;

/*Construir una consulta en SQL para obtener los nombres y apellidos de empleados y de clientes que cumplan que, 
para clientes que hayan hecho pedidos con el empleado, la dirección del empleado es alfabéticamente menor que la 
dirección del cliente, para los pedidos realizados en 2024*/
select emp.nombre, emp.apellidos, cli.nombre, cli.apellidos
from empleados emp join pedidos p on emp.idempleado = p.empleadoventa join clientes cli on cli.idcliente = p.idcliente
where cli.direccionpostal > emp.direccion and extract(year from p.fechapedido) = 2024;


--p2
/*Construir una consulta en SQL para obtener el idpedido, nombre y apellidos del cliente, nombre y apellidos del 
empleado y fecha de pedido y de entrega de los pedidos realizados en 2024 por empleados de trabajo "comercial" con 
una fecha de entrega superior a la fecha del pedido para clientes de Valencia.*/
select p.idpedido, cli.nombre, cli.apellidos, emp.nombre, emp.apellidos, p.fechapedido, p.fechaentrega
from clientes cli join pedidos p on cli.idcliente = p.idcliente join empleados emp on p.empleadoventa = emp.idempleado
where extract(year from p.fechapedido) = 2024 and upper(emp.trabajo) = 'comercial' and p.fechaentrega > p.fechapedido and upper(cli.ciudad) = 'Valencia';

/*Construir una consulta en SQL para obtener los datos completos de los clientes (sin duplicados) que hayan 
realizado pedidos que incluyan algún producto suministrado por proveedores cuyo nombre contenga la palabra "ANIMAL".
Se debe usar al menos una subconsulta.*/
select distinct * from clientes
where idcliente in (select idcliente from  pedidos p
    where p.idpedido in (select idpedido from pedidos_productos pp join
        productos prod on pp.idproducto = prod.idproducto join proveedores prov on prod.idproveedor = prov.idproveedor
        where upper(prov.nombre) like '%ANIMAL%'));

/*Construir una consulta en SQL para obtener el nombre completo del cliente, su ciudad y la cantidad de pedidos realizados 
para clientes que han hecho más de 5 pedidos.*/
select cli.nombre, cli.apellidos, cli.ciudad, count(p.idpedido)
from clientes cli join pedidos p on cli.idcliente = p.idcliente
group by cli.nombre, cli.apellidos, cli.idcliente, cli.ciudad
having count(p.idpedido) > 5;

--p3
/*Construir una sentencia SQL para crear una tabla (RESUMEN_CURSOS) que contenga un resumen de los cursos de 
los clientes por años, con los siguientes datos: nombre, apellidos e identificador del cliente, año, número 
de cursos realizados e importe total de los cursos, con las restricciones adecuadas. Construir una sentencia 
INSERT que permita rellenar la tabla con los datos existentes en la base de datos.*/
create table RESUMEN_CURSOS(
    nombre varchar2(20) not null,
    apellidos varchar2(50),
    idcliente number,
    anyo number,
    cursos_realizados number,
    importe_total float not null,

    constraint resumen_pk primary key(idcliente, anyo),
    constraint resumen_fk foreign key(idcliente) references clientes(idcliente)
)

insert into RESUMEN_CURSOS
select cli.nombre, cli.apellidos, cli.idcliente, extract(year from c.fechahora), count(cc.idcurso), sum(c.precio)
from clientes cli join cursos_cliente cc on cli.idcliente = cc.idcliente
join cursos c on cc.idcurso = c.idcurso
group by cli,nombre, cli.apellidos, cli.cliente, extract(year from c.fechahora);

/*Construir una sentencia SQL para añadir dos columnas que contengan la fecha del primer curso y la fecha del 
último curso en cada año (FINI, FFIN). Añadir una restricción a la tabla que haga que la fecha final sea igual o 
posterior a la fecha inicial.*/
alter table RESUMEN_CURSOS
add FINI DATE
add FFIN DATE
add constraint fecha_check check(FFIN >= FINI);

--p4
/*Construir un procedimiento PL/SQL que reciba como parámetro de entrada un identificador de cliente y un año 
y genere un informe por la salida estándar (DBMS_OUTPUT) que contenga una línea con el nombre, apellidos y email 
del cliente, una segunda línea con el número total de pedidos realizados por el cliente en el año que se pasa como 
parámetro, y una tercera línea con la cantidad de mascotas que tiene el cliente que hayan nacido en el año en 
cuestión.*/
create or replace procedure pr1(p_idcliente clientes.idcliente%type, p_año number) is
    v_nombre clientes.nombre%type;
    v_apellidos clientes.apellidos%type;
    v_email clientes.email%type;
    v_total_pedidos number;
    v_cantidad_mascotas number;
BEGIN
    select nombre, apellidos, email
    into v_nombre, v_apeellidos, v_email
    from clientes
    where cli.idcliente = p_idcliente;

    dbms_output.put_line('Nombre completo: ' || v_nombre || v_apellidos|| 'Email: '||v_email);

    select count(idpedido) into v_total_pedidos
    from pedidos 
    where idcliente = p_idcliente and extract(year from p.fechapedido) = p_año;

    dbms_output.put_line('Numero de pedidos: ' ||v_total_pedidos);

    select count(m.idmascota) into v_cantidad_mascotas
    from mascotas 
    where idcliente = p_idcliente and extract(year from m.fechanac) = p_año;

    dbms_output.put_line('Cantidad mascotas: ' ||v_cantidad_mascotas);
EXCEPTION
    when others then dbms_output.put_line('Error.');
END pr1;
    
/*Construir un procedimiento PL/SQL para modificar los datos de un pedido. 
El procedimiento debe modificar el precio de los productos del pedido (precioreal), 
incrementándolo en un 10%, para aquellos productos que tengan un stock inferior a 
una cantidad determinada. El procedimiento recibirá como entrada 2 parámetros: un 
identificador de pedido y un número entero que representa la cantidad de stock por 
debajo de la cual se debe incrementar el precio de los productos. Cada cambio realizado 
debe reportarse a la salida estándar (DBMS_OUTPUT), indicando el identificador de 
pedido, el identificador de producto al que se le incrementa el precio, y el precio final. 
Hacer un tratamiento adecuado de excepciones.*/
create or replace procedure p1(p_idpedido pedidos.idpedido%type, p_minimo number) IS
    cursor c1 IS
    select p.precioreal, pp.idpedido
    from productos p join pedidos_productos pp on p.idproducto = pp.idproducto
    where pp.idpedido = p_idpedido and p.stock < p_minimo;
BEGIN
    for reg_c1 in c1 loop
        update productos
        set precioreal = reg_c1.precioreal * 1.1
        where idpedido = p_idpedido and reg_c1.idproducto = idproducto;
        dbms_output.put_line('Pedido: '|| p_idpedido || 'Producto: '||reg_c1.idproducto|| 'Precio: '||(reg_c1.precioreal * 1.1));
    end loop;
    commit;
EXCEPTION
    when others THEN
    dbms_output.put_line('Error, deshaciendo cambios...');
    rollback;
END p1;


--gemini
create or replace procedure p(p_idcliente clientes.idcliente%type) is 
    cursor c1 is 
    select pp.idproducto, p.precioreal , pp.idpedido
    from productos_pedidos pp join pedidos p on pp.idproducto = p.idproducto
    join clientes cli on p.idcliente = cli.idcliente
    where p.precioreal > 50 and cli.idcliente = p_idcliente;

    v_num_mascotas number;
BEGIN
    select count(idmascota) into v_num_mascotas
    from mascotas 
    where idcliente = p_idcliente;

    if v_num_mascotas = 0 THEN
        dbms_output.put_line('El cliente no tiene mascotas, promoción no aplicable');
        return;
    end if;

    for reg_c1 in c1 loop
        update pedidos_productos
        set precioreal = precioreal * 0.95
        where idproducto = reg_c1.idproducto and idpedido = reg.idpedido;

        dbms_output.put_line('Pedido: ' ||reg_c1.idpedido|| 'Producto: '||reg_c1.idproducto|| 'Precio actualizado: '||(reg_c1.precioreal * 0.95));
    end loop;
        commit;
    EXCEPTION
        when OTHERS then 
        dbms_output.put_line('Error desaciendo cambios...');
    END p;

--boletin extra
create or replace function f_horas_reales(p_profesor gdi.npfirma%type, p_fecha date) return number is
    cursor c1 is 
    select hini, hfin from gdi
    where npifirma = p_profesor and fecha = p_fecha
    order by hini;

    v_horas_totales number := 0;
    v_hora_ini date;
    v_hora_fin date;
BEGIN
    for reg_c1 in c1 loop
        if v_hora_ini is null then 
            v_hora_ini := reg_c1.hini;
            v_hora_fin := reg_c1.hfi;

        elsif reg_c1.hini >= v_hora_fin THEN
            v_horas_total := v_horas_totales + ((v_hora_fin - v_hora_ini)*24);
            v_hora_ini := reg_c1.hini;
            v_hora_fin := reg_c1.hfin;
        else 
            if reg_c1.hfin > v_hora_fin THEN
                v_hora_fin := reg_c1.hfin;
            end if;
        end if;
    end loop;

    if v_hora_fin is not null then
        v_horas_totales := v_horas_totales + (((v_hora_fin - v_hora_ini)*24));
    end if;
    return v_horas_totales;
END f_horas_reales;

SELECT 
    NVL(NPIFIRMA, NPI) AS Profesor,
    FECHA,
    CASE 
        WHEN NPIFIRMA IS NULL THEN SUM((hfin - hini) * 24)
        ELSE MAX(f_horas_reales(NVL(NPIFIRMA, NPI), fecha))
    END AS horas_totales
FROM GDI
WHERE FECHA BETWEEN TO_DATE('01/09/2024', 'DD/MM/YYYY') AND TO_DATE('31/07/2025', 'DD/MM/YYYY')
GROUP BY NVL(NPIFIRMA, NPI), FECHA, NPIFIRMA;


















