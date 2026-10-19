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
	
	Error en la línea de comandos : 1 Columna : 8
	Informe de error -
	Error SQL: ORA-20001: Error: La longitud no puede ser negativa o demasiado grande
	ORA-06512: en "GIISGBD213.TGR_COMPROBAR_RIOS", línea 3
	ORA-04088: error durante la ejecución del disparador 'GIISGBD213.TGR_COMPROBAR_RIOS'

	https://docs.oracle.com/error-help/db/ora-20001/

	More Details :
	https://docs.oracle.com/error-help/db/ora-20001/
	https://docs.oracle.com/error-help/db/ora-06512/
	https://docs.oracle.com/error-help/db/ora-04088/
	
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
	
	
	
   