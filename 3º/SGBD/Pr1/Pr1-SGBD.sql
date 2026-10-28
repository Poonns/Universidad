--Ej1
--creacion tabla paises
create table PAISES(
    id_pais number,
    nombre varchar2(64) not null,
    continente varchar2(24) not null,
    capital varchar2(64) not null,
    situacion varchar2(128),
    notas varchar2(128),
    
    constraint pk_id_pais primary key (id_pais)
);

--inserts paises 
INSERT INTO PAISES (id_pais, nombre, continente, capital, situacion, notas) 
VALUES (1, 'España', 'Europa', 'Madrid', 'Sur de Europa', 'Península Ibérica');

INSERT INTO PAISES (id_pais, nombre, continente, capital, situacion, notas) 
VALUES (2, 'Portugal', 'Europa', 'Lisboa', 'Oeste de Europa', NULL);

INSERT INTO PAISES (id_pais, nombre, continente, capital, situacion, notas) 
VALUES (3, 'Francia', 'Europa', 'París', 'Europa Occidental', NULL);
	


--creacion tabla rios
create table RIOS(
    id_rio number,
    nombre varchar2(100) not null,
    longitud number,
    cuenca number,
    caudal_medio number,
    nacimiento varchar2(64) not null,
    pais_nacimiento number not null,
    desembocadura varchar2(64) not null,
    pais_desemboc number not null,
    
    constraint pk_id_rio primary key (id_rio),
    constraint fk_pais_nacimiento foreign key (pais_nacimiento) references PAISES(id_pais),
    constraint fk_pais_desemboc foreign key (pais_desemboc) references PAISES(id_pais)
);

--inserts rios
INSERT INTO RIOS (id_rio, nombre, longitud, cuenca, caudal_medio, nacimiento, pais_nacimiento, desembocadura, pais_desemboc) 
VALUES (1, 'Ebro', 930, 85362, 600, 'Fontibre', 1, 'Mar Mediterráneo', 1);

INSERT INTO RIOS (id_rio, nombre, longitud, cuenca, caudal_medio, nacimiento, pais_nacimiento, desembocadura, pais_desemboc) 
VALUES (2, 'Tajo', 990, 80600, 444, 'Sierra de Albarracín', 1, 'Océano Atlántico', 2);

INSERT INTO RIOS (id_rio, nombre, longitud, cuenca, caudal_medio, nacimiento, pais_nacimiento, desembocadura, pais_desemboc) 
VALUES (3, 'Duero', 897, 98073, 700, 'Picos de Urbión', 1, 'Océano Atlántico', 2);

INSERT INTO RIOS (id_rio, nombre, longitud, cuenca, caudal_medio, nacimiento, pais_nacimiento, desembocadura, pais_desemboc) 
VALUES (4, 'Sena', 777, 79000, 500, 'Meseta de Langres', 3, 'Canal de la Mancha', 3);

--Ej2
--Trigger
create or replace trigger tgr_comprobar_rios 
before insert or update
on RIOS
for each row
begin
    if :new.longitud < 0 or :new.longitud > 10000 then
        raise_application_error(-20001, 'Error: La longitud no puede ser negativa o demasiado grande');
end if;
    if :new.caudal_medio < 0 or :new.caudal_medio > 10000 then
        raise_application_error(-20002, 'Error: El caudal no puede ser negativa o demasiado grande');
end if;
end;
/

	--a)
	select count(*) from RIOS;
	--count: 4
	
	--b)
	--inserción de un rio
	INSERT INTO RIOS (id_rio, nombre, longitud, cuenca, caudal_medio, nacimiento, pais_nacimiento, desembocadura, pais_desemboc) 
	VALUES (5, 'Garona', 529, 84811, 631, 'Pirineos', 1, 'Océano Atlántico', 3);
	
	--c)
	update RIOS 
	set longitud = -1 
	where id_rio = 1;
	
	--Informe de error -
	--Error SQL: ORA-20001: Error: La longitud no puede ser negativa o demasiado grande
	--ORA-06512: en "GIISGBD213.TGR_COMPROBAR_RIOS", línea 3
	--ORA-04088: error durante la ejecución del disparador 'GIISGBD213.TGR_COMPROBAR_RIOS'

	--https://docs.oracle.com/error-help/db/ora-20001/

	--More Details :
	--https://docs.oracle.com/error-help/db/ora-20001/
	--https://docs.oracle.com/error-help/db/ora-06512/
	--https://docs.oracle.com/error-help/db/ora-04088/
	
	--d)
	alter trigger tgr_comprobar_rios disable;
	
	--e)
	--actualizacion fila
	update RIOS 
	set longitud = -1 
	where id_rio = 1;

	--mostramos el cambio
	select * from RIOS where id_rio = 1;
	
	--f)
	alter trigger tgr_comprobar_rios enable;
	
	--g)
	rollback;
	
	--h)
	select count(*) from RIOS;
	--count: 5
	--tendriamos el mismo resultado que en el apartado a) porque hemos deshecho los cambios con rollback, pero hemos insertado un rio mas en el apartado b)
	
--Ej3
--Trigger 
create or replace trigger tgr_pais_desemb
before insert or update
on RIOS
for each row 
declare 
    v_continente varchar(30);
begin
    select continente into v_continente 
    from PAISES
    where id_pais = :new.pais_desemboc;
    
    if v_continente != 'Europa' then 
        raise_application_error(-20003, 'Error: El rio debe desembocar en Europa');
    end if;
end;
--Trigger TGR_PAIS_DESEMB compilado

--insertamos un pais que no desemboque en Europa para hacer la prueba
INSERT INTO PAISES (id_pais, nombre, continente, capital, situacion, notas) 
VALUES (4, 'Egipto', 'África', 'El Cairo', 'Norte de África', NULL);
--1 fila insertadas

--insertamos un rio nuevo asignandole un egipto como pais de desembocadura
INSERT INTO RIOS (id_rio, nombre, longitud, cuenca, caudal_medio, nacimiento, pais_nacimiento, desembocadura, pais_desemboc) 
VALUES (6, 'Río Falso', 100, 5000, 100, 'Desierto', 4, 'Mar Rojo', 4);

--Informe de error -
--Error SQL: ORA-20003: Error: El rio debe desembocar en Europa
--ORA-06512: en "GIISGBD213.TGR_PAIS_DESEMB", línea 9
--ORA-04088: error durante la ejecución del disparador 'GIISGBD213.TGR_PAIS_DESEMB'

--https://docs.oracle.com/error-help/db/ora-20003/

--More Details :
--https://docs.oracle.com/error-help/db/ora-20003/
--https://docs.oracle.com/error-help/db/ora-06512/
--https://docs.oracle.com/error-help/db/ora-04088/

--intentamos actualizar la desembocadura de un rio ya guardado
update RIOS 
set pais_desemboc = 4 
where id_rio = 1;

--Informe de error -
--Error SQL: ORA-20003: Error: El rio debe desembocar en Europa
--ORA-06512: en "GIISGBD213.TGR_PAIS_DESEMB", línea 9
--ORA-04088: error durante la ejecución del disparador 'GIISGBD213.TGR_PAIS_DESEMB'

--https://docs.oracle.com/error-help/db/ora-20003/

--More Details :
--https://docs.oracle.com/error-help/db/ora-20003/
--https://docs.oracle.com/error-help/db/ora-06512/
--https://docs.oracle.com/error-help/db/ora-04088/

--Ej4
--creamos la nueva columna
alter table PAISES add num_rios number default 0;
--Table PAISES alterado.

--actualizamos los datos existentes
update PAISES p
set num_rios = (select count(*) from RIOS r where r.pais_desemboc = p.id_pais);

--Trigger 
create or replace trigger trg_actualizar_num_rios
after insert or update OF pais_desemboc or delete on RIOS
for each row
begin
    if inserting then
        --si insertamos un río sumamos 1 al país de desembocadura
        update PAISES 
        set num_rios = num_rios + 1 
        where id_pais = :new.pais_desemboc;
        
    elsif deleting then
        --si borramos un río restamos 1 al país afectado
        update PAISES 
        set num_rios = num_rios - 1 
        where id_pais = :old.pais_desemboc;
        
    elsif updating then
        --restamos 1 al país antiguo
        update PAISES 
        set num_rios = num_rios - 1 
        where id_pais = :old.pais_desemboc;
            
        --sumamos 1 al país nuevo
        update PAISES 
        set num_rios = num_rios + 1 
        where id_pais = :new.pais_desemboc;
    end if;
end;

--Trigger TRG_ACTUALIZAR_NUM_RIOS compilado

--comprobamos cuantos rios tienen 
select id_pais, nombre, num_rios 
from PAISES where id_pais in (1, 2);

--ID_PAIS	NOMBRE	NUM_RIOS
--1	España	1
--2	Portugal	2

--insertamos un rio de prueba
INSERT INTO RIOS (id_rio, nombre, longitud, cuenca, caudal_medio, nacimiento, pais_nacimiento, desembocadura, pais_desemboc) 
VALUES (99, 'Río Prueba', 10, 10, 10, 'Montaña', 1, 'Mar', 1);
--1 fila insertadas
select id_pais, nombre, num_rios from PAISES where id_pais in (1, 2);

--ID_PAIS	NOMBRE	NUM_RIOS	
--1	España	2
--2	Portugal	2

--cambiamos la desembocadura
update RIOS 
set pais_desemboc = 2 where id_rio = 99;
--1 fila actualizadas
select id_pais, nombre, num_rios from PAISES where id_pais in (1, 2);

--ID_PAIS	NOMBRE	NUM_RIOS	
--1	España	1
--2	Portugal	3

--borramos el rio de prueba
delete from RIOS where id_rio = 99;
select id_pais, nombre, num_rios from PAISES where id_pais in (1, 2);

--ID_PAIS	NOMBRE	NUM_RIOS	
--1	España	1
--2	Portugal	2

--deshacemos los cambios
rollback;

--comprobamos que todo haya quedado igual
select id_pais, nombre, num_rios from PAISES where id_pais in (1, 2);

--ID_PAIS	NOMBRE	NUM_RIOS
--1	España	1
--2	Portugal	2

--Ej5
create table CAMBIOLOG (
    id number generated always as identity,
    nombre_tabla varchar2(30),
    fecha_cambio date,
    constraint pk_cambiolog primary key (id)
);

--Trigger paises
create or replace trigger trg_log_paises
after insert or delete or update on PAISES
for each row
declare
    v_total number;
begin
    --registramos el cambio
    INSERT INTO CAMBIOLOG (nombre_tabla, fecha_cambio) 
    VALUES ('PAISES', sysdate);
    
    --contamos los registros
    select count(*) into v_total from CAMBIOLOG;
    
    --si pasamos de 10 borramos la fecha minima
    if v_total > 10 then
        delete from CAMBIOLOG 
        where id = (select min(id) from CAMBIOLOG);
    end if;
end;

--Trigger rios
create or replace trigger trg_log_rios
after insert or delete or update on RIOS
for each row
declare
    v_total number;
begin
    --registramos el cambio
    INSERT INTO CAMBIOLOG (nombre_tabla, fecha_cambio) 
    VALUES ('RIOS', SYSDATE);
    
    --contamos los registros
    select count(*) into v_total from CAMBIOLOG;
    
    --si pasamos de 10 borramos la fecha minima
    if v_total > 10 then
        delete from CAMBIOLOG 
        where id = (select min(id) from CAMBIOLOG);
    end if;
end;

--forzamos cambios en RIOS
INSERT INTO RIOS (id_rio, nombre, longitud, cuenca, caudal_medio, nacimiento, pais_nacimiento, desembocadura, pais_desemboc) 
VALUES (99, 'Río Log', 10, 10, 10, 'Montaña', 1, 'Mar', 1);

update RIOS 
set longitud = 20 where id_rio = 99;

delete from RIOS where id_rio = 99;

--forzamos cambios en PAISES
update PAISES 
set nombre = 'España mod' where id_pais = 1;

--comprobamos historial
select * from CAMBIOLOG;

--NOMBRE_TABLA	FECHA_CAMBIO
--RIOS		22/10/26
--RIOS		22/10/26
--RIOS		22/10/26


--limpiamos la BBDD
rollback;

--Ej6
--Trigger error
create or replace trigger trg_error_mutante
before update on RIOS
for each row
declare
    v_media_longitud number;
begin
    --error intencionado
    select avg(longitud) into v_media_longitud from RIOS;
end;

--forzamos el fallo
update RIOS 
set longitud = 1500 where id_rio = 1;

--Informe de error -
--Error SQL: ORA-04091: la tabla GIISGBD213.RIOS está mutando, puede que el disparador/la función no puedan verla
--ORA-06512: en "GIISGBD213.TRG_ERROR_MUTANTE", línea 5
--ORA-04088: error durante la ejecución del disparador 'GIISGBD213.TRG_ERROR_MUTANTE'

--https://docs.oracle.com/error-help/db/ora-04091/04091. 00000 -  "table %s.%s is mutating, trigger/function may not see it"
--*Cause:    A trigger (or a user defined PL/SQL function that is referenced in
--          this statement) attempted to look at (or modify) a table that was
--           in the middle of being modified by the statement which fired it.
--*Action:   Rewrite the trigger (or function) so it does not read that table.
--
--More Details :
--https://docs.oracle.com/error-help/db/ora-04091/
--https://docs.oracle.com/error-help/db/ora-06512/
--https://docs.oracle.com/error-help/db/ora-04088/

--Borramos el trigger
drop trigger trg_error_mutante;

--Ej7
--Trigger de PAISES
create or replace trigger trg_pk_paises
before insert on PAISES
for each row
declare
    v_siguiente_id number;
begin
    --si el id es nulo calcula el siguiente
    if :new.id_pais is null then
        select nvl(max(id_pais), 0) + 1 into v_siguiente_id from PAISES;
        :new.id_pais := v_siguiente_id;
    end if;
end;

--Trigger de RIOS
create or replace trigger trg_pk_rios
before insert on RIOS
for each row
declare
    v_siguiente_id number;
begin
    if: new.id_rio is null then
        select nvl(max(id_rio), 0) + 1 into v_siguiente_id from RIOS;
        :new.id_rio := v_siguiente_id;
   end if;
end;

--insertamos un pais omitiendo la primera columna y comprobamos
INSERT INTO PAISES (nombre, capital, continente, num_rios) 
VALUES ('Kosovo', 'Pristina', 'Europa', 0);

select * from PAISES where nombre = 'Kosovo';

--ID_PAIS	NOMBRE	CONTINENTE	CAPITAL	SITUACION	NOTAS	NUM_RIOS
--5	Kosovo	Europa	Pristina	(null)	(null)	0

rollback;

--Ej8
--Trigger A
create or replace trigger trg_a
before update on PAISES
for each row
begin
    INSERT INTO CAMBIOLOG (nombre_tabla, fecha_cambio) 
    VALUES ('Ejecutado Trigger A', sysdate);
end;

--Trigger B
create or replace trg_b
before update on PAISES
for each row
FOLLOWS trg_orden_a --indicamos que se espere a que termine A
begin
    INSERT INTO CAMBIOLOG (nombre_tabla, fecha_cambio) 
    VALUES ('Ejecutado Trigger B', sysdate);
end;

--Alternativa: hacer un unico trigger con el orden de las instrucciones en las que queremos que se ejecuten

--modificamos cualquier pais para hacer la prueba
update PAISES 
set continente = 'Europa' where id_pais = 1;

--comprobamos el orden
select * from CAMBIOLOG;

--NOMBRE_TABLA	FECHA_CAMBIO
--Ejecutado Trigger A	22/10/26
--Ejecutado Trigger B	22/10/26

















	
	
	
   