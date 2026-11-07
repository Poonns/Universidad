--Ej1
--1.1
--creamos las tablas (ejecutar crea_tablas.sql)
--Table ACTORES creado.
--Table PELICULAS creado.
--Table ACTUACION creado.

--una vez creadas metemos la informacion en ellas:
insert into actores select * from GIISGBD.ACTORES_DATOS2;
insert into peliculas select * from GIISGBD.PELICULAS_DATOS2;
insert into actuacion select * from GIISGBD.ACTUACION_DATOS2;
commit;

--1.2
--creamos los indices sobre la columna sexo y anyo
create index idx_sexo on actores (sexo);
create index idx_anyo on peliculas (anyo);

--1.3 
--generamos las estadisticas completas de las tablas incluyendo los indices creados
exec dbms_stats.gather_table_stats (ownname => 'GIISGBD213', tabname => 'ACTORES', cascade => true);
exec dbms_stats.gather_table_stats (ownname => 'GIISGBD213',tabname => 'PELICULAS', cascade => true);
exec dbms_stats.gather_table_stats (ownname => 'GIISGBD213', tabname => 'ACTUACION', cascade => true);

--1.4
--consultamos la informacion de las estadisticas
	--a)tablas
	select table_name, num_rows, blocks, avg_row_len 
	from user_tables 
	where table_name in ('ACTORES', 'PELICULAS', 'ACTUACION');
	
	--ACTORES	9146	244	108
	--ACTUACION	12500	244	111
	--PELICULAS	11182	370	210
	
	--b)columnas
	select table_name, column_name, num_distinct, low_value, high_value, num_nulls, avg_col_len, histogram 
	from user_tab_columns 
	where table_name in ('ACTORES', 'PELICULAS', 'ACTUACION');
	
	--ACTORES	OID	9146	C3173202	C33B3C35	0	5	NONE
	--ACTORES	NOM	9146	41616C746F2C20416E75202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020	6C6120507565626C612C204E69C3B161206465202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020	0	101	NONE
	--ACTORES	SEXO	1	4D	4D	0	2	NONE
	--ACTUACION	ACTOR	9146	C3173202	C33B3C35	0	5	NONE
	--ACTUACION	PELI	11182	C3173108	C33B3B54	0	5	NONE
	--ACTUACION	PAPEL	6585	317374204174746F726E657920202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020	C3897661202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020	0	101	NONE
	--PELICULAS	OID	11182	C3173108	C33B3B54	0	5	NONE
	--PELICULAS	TITULO	11014	2431303030206120546F756368646F776E2020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020	C3BF72626F6C2064652045737061C3B1612C20456C20202020202020202020202020202020202020202020202020202020202020202020202020202020202020	0	201	NONE
	--PELICULAS	ANYO	95	80	C21464	0	4	NONE
	
	--c)indices
	select index_name, table_name, blevel, leaf_blocks, distinct_keys, avg_leaf_blocks_per_key, avg_data_blocks_per_key, clustering_factor, num_rows 
	from user_indexes 
	where index_name in ('IDX_SEXO', 'IDX_ANYO');
	
	--IDX_ANYO	PELICULAS	1	24	95	1	88	8450	11182
	--IDX_SEXO	ACTORES	1	17	1	17	143	143	9146

--1.5
--ACTUACION tiene mas filas que ACTORES pero tienen el mismo numero de bloques (oracle reserva el espacio en el disco en bloques fijos)
--En la avg_col de TITULO, PAPEL y NOM ocupan 201 y 101 bytes, en el documento se especificaba que las tablas se creaban usando char(200) y char(100). Lo que indica esto es que aunque tu alguno de esos campos los rellenes con 2 letras, Oracle rellenara el resto con espacios en blanco derochando memoria
--Solo hay 1 num_distc en sexo, esto quiere decir que todos los valores que se han introducido tienen el mismo sexo, por tanto no tiene sentido tener un indice sobre este campos

--1.6
--eliminamos los indices
drop index idx_sexo;
drop index idx_anyo;

--eliminamos las estadisticas
exec dbms_stats.delete_schema_stats (ownname => 'GIISGBD213');

--1.7
--ejecutamos las consultas del apartado 4 y comprobamos
select table_name, num_rows, blocks, avg_row_len 
from user_tables 
where table_name in ('ACTORES', 'PELICULAS', 'ACTUACION');

--ACTORES	(null)	(null)	(null)		
--ACTUACION	(null)	(null)	(null)	
--PELICULAS	(null)	(null)	(null)

select table_name, column_name, num_distinct, low_value, high_value, num_nulls, avg_col_len, histogram 
from user_tab_columns 
where table_name in ('ACTORES', 'PELICULAS', 'ACTUACION');

--ACTORES	OID	(null)	(null)	(null)	(null)	(null)	NONE
--ACTORES	NOM	(null)	(null)	(null)	(null)	(null)	NONE
--ACTORES	SEXO	(null)	(null)	(null)	(null)	(null)	NONE
--ACTUACION	ACTOR	(null)	(null)	(null)	(null)	(null)	NONE
--ACTUACION	PELI	(null)	(null)	(null)	(null)	(null)	NONE
--ACTUACION	PAPEL	(null)	(null)	(null)	(null)	(null)	NONE
--PELICULAS	OID	(null)	(null)	(null)	(null)	(null)	NONE
--PELICULAS	TITULO	(null)	(null)	(null)	(null)	(null)	NONE
--PELICULAS	ANYO	(null)	(null)	(null)	(null)	(null)	NONE

select index_name, table_name, blevel, leaf_blocks, distinct_keys, avg_leaf_blocks_per_key, avg_data_blocks_per_key, clustering_factor, num_rows 
from user_indexes 
where index_name in ('IDX_SEXO', 'IDX_ANYO');

--IDX_ANYO	PELICULAS	(null)	(null)	(null)	(null)	(null)	(null)	(null)
--IDX_SEXO	ACTORES	(null)	(null)	(null)	(null)	(null)	(null)	(null)

--Ej2
--al inicio de cada ejercicio vaciamos las tablas y volvemos a cargar los datos para asegurarnos de que se encuentran ahi
DELETE FROM actuacion;
DELETE FROM peliculas;
DELETE FROM actores;

insert into actores select * from GIISGBD.ACTORES_DATOS2;
insert into peliculas select * from GIISGBD.PELICULAS_DATOS2;
insert into actuacion select * from GIISGBD.ACTUACION_DATOS2;
--2.1
--ejecutamos las consultas 
select *
from peliculas
where anyo > 1992;
--1400 filas recuperadas en 1,164s

select *
from peliculas p, actuacion act
where p.oid = act.peli
and anyo > 1992;
--50 filas recuperadas en 0,111s

select p.*
from actores a, actuacion x, peliculas p
where a.NOM = 'Loy, Myrna'
and a.OID = X.ACTOR
and p.OID = X.PELI;
--11 fils recuperadas en 0,089s

--2.2
--aplicamos el explain plan for (analiza la consulta y calcula el plan de ejecucion) y hacemos el display (dibuja el plan)
explain plan for
select * from peliculas where ANYO > 1992;
select * from table(dbms_xplan.display);

--Plan hash value: 2378278331
 
---------------------------------------------------------------------------------
--| Id  | Operation         | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |           |  1979 |   440K|   102   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| PELICULAS |  1979 |   440K|   102   (0)| 00:00:01 |
---------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
---------------------------------------------------
 
--   1 - filter("ANYO">1992)
 
--Note
-----
--   - dynamic statistics used: dynamic sampling (level=2)
--   - SQL plan baseline "SQL_PLAN_37ncmpadunta7d6672d41" used for this statement

explain plan for 
select * from peliculas p, actuacion act where p.OID = act.PELI and ANYO > 1992;
select * from table(dbms_xplan.display);

--Plan hash value: 1777033210
 
-----------------------------------------------------------------------------------------
--| Id  | Operation             | Name          | Rows  | Bytes | Cost (%CPU)| Time     |
-----------------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT      |               |  2420 |   841K|   204   (0)| 00:00:01 |
--|*  1 |  HASH JOIN            |               |  2420 |   841K|   204   (0)| 00:00:01 |
--|*  2 |   TABLE ACCESS FULL   | PELICULAS     |  1979 |   440K|   102   (0)| 00:00:01 |
--|   3 |   INDEX FAST FULL SCAN| SYS_C00465413 | 12006 |  1500K|   102   (0)| 00:00:01 |
-----------------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - access("P"."OID"="ACT"."PELI")
--   2 - filter("ANYO">1992)
 
--Note
-------
--   - dynamic statistics used: dynamic sampling (level=2)
--   - SQL plan baseline "SQL_PLAN_5rbfp05f7uk3m7c4a039a" used for this statement

explain plan for
select p.* from actores a, actuacion x, peliculas p
where a.NOM = 'Loy, Myrna' and a.OID = X.ACTOR and p.OID = X.PELI;
select * from table(dbms_xplan.display);

--Plan hash value: 1023209607
 
------------------------------------------------------------------------------------------------
--| Id  | Operation                    | Name          | Rows  | Bytes | Cost (%CPU)| Time     |
------------------------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT             |               |     9 |  3321 |    80   (0)| 00:00:01 |
--|   1 |  NESTED LOOPS                |               |     9 |  3321 |    80   (0)| 00:00:01 |
--|   2 |   NESTED LOOPS               |               |     9 |  3321 |    80   (0)| 00:00:01 |
--|   3 |    NESTED LOOPS              |               |     9 |  1269 |    71   (0)| 00:00:01 |
--|*  4 |     TABLE ACCESS FULL        | ACTORES       |     3 |   345 |    68   (0)| 00:00:01 |
--|*  5 |     INDEX RANGE SCAN         | SYS_C00465413 |     3 |    78 |     1   (0)| 00:00:01 |
--|*  6 |    INDEX UNIQUE SCAN         | SYS_C00465409 |     1 |       |     0   (0)| 00:00:01 |
--|   7 |   TABLE ACCESS BY INDEX ROWID| PELICULAS     |     1 |   228 |     1   (0)| 00:00:01 |
------------------------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   4 - filter("A"."NOM"='Loy, Myrna')
--   5 - access("A"."OID"="X"."ACTOR")
--   6 - access("P"."OID"="X"."PELI")
 
--Note
-------
--   - dynamic statistics used: dynamic sampling (level=2)

	--a)Sí, los números de la columna Rows son inexactos (estima 12.006 actuaciones en lugar de las 12.500 reales) porque Oracle no tiene las estadísticas para calcularlo bien
	
--2.3
--el mensaje significa que al haber borrado las estadísticas, Oracle ha tenido que leer rápidamente unos pocos bloques al azar para intentar estimar los datos a ciegas

--2.4
--creamos las estadisticas de las tablas (igual que antes pero en vez de table ponemos schema para que lo con todas las tablas de golpe)
exec dbms_stats.gather_schema_stats(ownname => user, cascade => true);

--Procedimiento PL/SQL terminado correctamente.
--Explicado.
--Explicado.
-->>Query Run In:Resultado de la Consulta
--Explicado.
-->>Query Run In:Resultado de la Consulta 1
--Procedimiento PL/SQL terminado correctamente.


--2.5
--volvemos a ejecutar las consultas
explain plan for
select * from peliculas where ANYO > 1992;
select * from table(dbms_xplan.display);

--Plan hash value: 2378278331
 
---------------------------------------------------------------------------------
--| Id  | Operation         | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |           |  1999 |   409K|   102   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| PELICULAS |  1999 |   409K|   102   (0)| 00:00:01 |
---------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - filter("ANYO">1992)
 
--Note
-------
--   - SQL plan baseline "SQL_PLAN_37ncmpadunta7d6672d41" used for this statement

explain plan for 
select * from peliculas p, actuacion act where p.OID = act.PELI and ANYO > 1992;
select * from table(dbms_xplan.display);

--Plan hash value: 1777033210
 
-----------------------------------------------------------------------------------------
--| Id  | Operation             | Name          | Rows  | Bytes | Cost (%CPU)| Time     |
-----------------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT      |               |  2235 |   700K|   186   (0)| 00:00:01 |
--|*  1 |  HASH JOIN            |               |  2235 |   700K|   186   (0)| 00:00:01 |
--|*  2 |   TABLE ACCESS FULL   | PELICULAS     |  1999 |   409K|   102   (0)| 00:00:01 |
--|   3 |   INDEX FAST FULL SCAN| SYS_C00465413 | 12500 |  1354K|    84   (0)| 00:00:01 |
-----------------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - access("P"."OID"="ACT"."PELI")
--   2 - filter("ANYO">1992)
 
--Note
-------

explain plan for
select p.* from actores a, actuacion x, peliculas p
where a.NOM = 'Loy, Myrna' and a.OID = X.ACTOR and p.OID = X.PELI;
select * from table(dbms_xplan.display);

--Plan hash value: 1023209607
 
------------------------------------------------------------------------------------------------
--| Id  | Operation                    | Name          | Rows  | Bytes | Cost (%CPU)| Time     |
------------------------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT             |               |     1 |   326 |    70   (0)| 00:00:01 |
--|   1 |  NESTED LOOPS                |               |     1 |   326 |    70   (0)| 00:00:01 |
--|   2 |   NESTED LOOPS               |               |     1 |   326 |    70   (0)| 00:00:01 |
--|   3 |    NESTED LOOPS              |               |     1 |   116 |    69   (0)| 00:00:01 |
--|*  4 |     TABLE ACCESS FULL        | ACTORES       |     1 |   106 |    68   (0)| 00:00:01 |
--|*  5 |     INDEX RANGE SCAN         | SYS_C00465413 |     1 |    10 |     1   (0)| 00:00:01 |
--|*  6 |    INDEX UNIQUE SCAN         | SYS_C00465409 |     1 |       |     0   (0)| 00:00:01 |
--|   7 |   TABLE ACCESS BY INDEX ROWID| PELICULAS     |     1 |   210 |     1   (0)| 00:00:01 |
------------------------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   4 - filter("A"."NOM"='Loy, Myrna')
--   5 - access("A"."OID"="X"."ACTOR")
--   6 - access("P"."OID"="X"."PELI")
 
--Note
-------
--   - SQL plan baseline "SQL_PLAN_4z5x6suxn6hpc21f1959c" used for this statement
	
	--a)Desaparece el mensaje de muestreo dinámico. Al tener estadísticas reales, las estimaciones de filas se vuelven exactas (muestra las 12.500 actuaciones reales y 1 sola fila para Myrna Loy) y el coste general calculado se reduce

--borramos las estadisticas
exec dbms_stats.delete_schema_stats(ownname => 'GIISGBD213');

--Ej3
DELETE FROM actuacion;
DELETE FROM peliculas;
DELETE FROM actores;

insert into actores select * from GIISGBD.ACTORES_DATOS2;
insert into peliculas select * from GIISGBD.PELICULAS_DATOS2;
insert into actuacion select * from GIISGBD.ACTUACION_DATOS2;
--3.1
explain plan for
select * from peliculas where ANYO=1920;
select * from table(dbms_xplan.display);

--Plan hash value: 2378278331
 
---------------------------------------------------------------------------------
--| Id  | Operation         | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |           |    47 |  9870 |   102   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| PELICULAS |    47 |  9870 |   102   (0)| 00:00:01 |
---------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - filter("ANYO"=1920)

explain plan for
select * from peliculas where ANYO<1920;
select * from table(dbms_xplan.display);

--Plan hash value: 2378278331
 
---------------------------------------------------------------------------------
--| Id  | Operation         | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |           |   393 | 89604 |   102   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| PELICULAS |   393 | 89604 |   102   (0)| 00:00:01 |
---------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - filter("ANYO"<1920)
 
--Note
-------
--   - dynamic statistics used: dynamic sampling (level=2)

explain plan for
select * from peliculas where ANYO>1920;
select * from table(dbms_xplan.display);

--Plan hash value: 2378278331
 
---------------------------------------------------------------------------------
--| Id  | Operation         | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |           | 10800 |  2404K|   102   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| PELICULAS | 10800 |  2404K|   102   (0)| 00:00:01 |
---------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - filter("ANYO">1920)
 
--Note
-------
--   - dynamic statistics used: dynamic sampling (level=2)

explain plan for
select * from peliculas where ANYO<>1920;
select * from table(dbms_xplan.display);

--Plan hash value: 2378278331
 
---------------------------------------------------------------------------------
--| Id  | Operation         | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |           | 11194 |  2492K|   102   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| PELICULAS | 11194 |  2492K|   102   (0)| 00:00:01 |
---------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - filter("ANYO"<>1920)
 
--Note
-------
--   - dynamic statistics used: dynamic sampling (level=2)

--3.2
--creamos las estadisticas y volvemos a obtener el plan de ejecucion
exec dbms_stats.gather_schema_stats(ownname => 'GIISGBD213', cascade => true);



explain plan for
select * from peliculas where ANYO=1920;
select * from table(dbms_xplan.display);

--Plan hash value: 2378278331
 
---------------------------------------------------------------------------------
--| Id  | Operation         | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |           |    47 |  9870 |   102   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| PELICULAS |    47 |  9870 |   102   (0)| 00:00:01 |
---------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--  1 - filter("ANYO"=1920)

explain plan for
select * from peliculas where ANYO<1920;
select * from table(dbms_xplan.display);

--Plan hash value: 2378278331
 
---------------------------------------------------------------------------------
--| Id  | Operation         | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |           |   450 | 94500 |   102   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| PELICULAS |   450 | 94500 |   102   (0)| 00:00:01 |
---------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - filter("ANYO"<1920)

explain plan for
select * from peliculas where ANYO>1920;
select * from table(dbms_xplan.display);

--Plan hash value: 2378278331
 
---------------------------------------------------------------------------------
--| Id  | Operation         | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |           | 10686 |  2191K|   102   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| PELICULAS | 10686 |  2191K|   102   (0)| 00:00:01 |
---------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - filter("ANYO">1920)

explain plan for
select * from peliculas where ANYO<>1920;
select * from table(dbms_xplan.display);

--Plan hash value: 2378278331
 
---------------------------------------------------------------------------------
--| Id  | Operation         | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |           | 11135 |  2283K|   102   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| PELICULAS | 11135 |  2283K|   102   (0)| 00:00:01 |
---------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - filter("ANYO"<>1920)

	--a)no, la estructura del plan de ejecucion es la misma, las cuatro consultas siguen haciendo un escaneo completo de la tabla ya que todavía no existe ningun indice sobre la columna anyo 
	
--3.3
--creamos el indice y generamos las estadisticas
create index idx_anyo on PELICULAS(ANYO);
exec dbms_stats.gather_table_stats(ownname => 'GIISGBD213', tabname => 'PELICULAS', cascade => true);

explain plan for
select * from peliculas where ANYO=1920;
select * from table(dbms_xplan.display);

--Plan hash value: 717746877
 
---------------------------------------------------------------------------------------------------
--| Id  | Operation                           | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT                    |           |    47 |  9870 |    37   (0)| 00:00:01 |
--|   1 |  TABLE ACCESS BY INDEX ROWID BATCHED| PELICULAS |    47 |  9870 |    37   (0)| 00:00:01 |
--|*  2 |   INDEX RANGE SCAN                  | IDX_ANYO  |    47 |       |     1   (0)| 00:00:01 |
---------------------------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   2 - access("ANYO"=1920)

explain plan for
select * from peliculas where ANYO<1920;
select * from table(dbms_xplan.display);

--Plan hash value: 2378278331
 
---------------------------------------------------------------------------------
--| Id  | Operation         | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |           |   450 | 94500 |   102   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| PELICULAS |   450 | 94500 |   102   (0)| 00:00:01 |
---------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - filter("ANYO"<1920)

explain plan for
select * from peliculas where ANYO>1920;
select * from table(dbms_xplan.display);

--Plan hash value: 2378278331
 
---------------------------------------------------------------------------------
--| Id  | Operation         | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |           | 10686 |  2191K|   102   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| PELICULAS | 10686 |  2191K|   102   (0)| 00:00:01 |
---------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - filter("ANYO">1920)

explain plan for
select * from peliculas where ANYO<>1920;
select * from table(dbms_xplan.display);

--Plan hash value: 2378278331
 
---------------------------------------------------------------------------------
--| Id  | Operation         | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |           | 11135 |  2283K|   102   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| PELICULAS | 11135 |  2283K|   102   (0)| 00:00:01 |
---------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - filter("ANYO"<>1920)

	--a)El optimizador solo usa el índice en la consulta ANYO=1920 porque recupera un porcentaje muy pequeño de filas, siendo la opción más barata
	
--3.4
--maximo donde usa el indice
explain plan for 
select * from PELICULAS where ANYO < 1914; 
select * from table(dbms_xplan.display);

--Plan hash value: 717746877
 
---------------------------------------------------------------------------------------------------
--| Id  | Operation                           | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT                    |           |   129 | 27090 |   100   (0)| 00:00:01 |
--|   1 |  TABLE ACCESS BY INDEX ROWID BATCHED| PELICULAS |   129 | 27090 |   100   (0)| 00:00:01 |
--|*  2 |   INDEX RANGE SCAN                  | IDX_ANYO  |   129 |       |     2   (0)| 00:00:01 |
---------------------------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   2 - access("ANYO"<1914)

--minimo donde usa el acceso directo a la tabla
explain plan for 
select * from PELICULAS where ANYO < 1915;
select * from table(dbms_xplan.display);

--Plan hash value: 2378278331
 
---------------------------------------------------------------------------------
--| Id  | Operation         | Name      | Rows  | Bytes | Cost (%CPU)| Time     |
---------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |           |   186 | 39060 |   102   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| PELICULAS |   186 | 39060 |   102   (0)| 00:00:01 |
---------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - filter("ANYO"<1915)

--3.5
--eliminamos el indice
drop index idx_anyo;

--Ej4
DELETE FROM actuacion;
DELETE FROM peliculas;
DELETE FROM actores;

insert into actores select * from GIISGBD.ACTORES_DATOS1;
insert into peliculas select * from GIISGBD.PELICULAS_DATOS1;
insert into actuacion select * from GIISGBD.ACTUACION_DATOS1;

--4.1
--creamos las estadisticas adecuadas
exec dbms_stats.gather_table_stats(ownname => 'GIISGBD213', tabname => 'ACTORES', cascade => true);

--4.2
--obtenemos el plan de ejecucion
explain plan for
select * from ACTORES where SEXO = 'H';

--Plan hash value: 2765877494
 
-------------------------------------------------------------------------------
--| Id  | Operation         | Name    | Rows  | Bytes | Cost (%CPU)| Time     |
-------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |         |  1172 |   123K|    68   (0)| 00:00:01 |
--|*  1 |  TABLE ACCESS FULL| ACTORES |  1172 |   123K|    68   (0)| 00:00:01 |
-------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   1 - filter("SEXO"='H')
 
--Note
-------
--   - SQL plan baseline "SQL_PLAN_5sagc315vquw0c4135987" used for this statementselect * from table(dbms_xplan.display);

explain plan for
select count(*) from ACTORES where SEXO = 'H';
select * from table(dbms_xplan.display);

--Plan hash value: 3398582430
 
--------------------------------------------------------------------------------
--| Id  | Operation          | Name    | Rows  | Bytes | Cost (%CPU)| Time     |
--------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT   |         |     1 |     2 |    68   (0)| 00:00:01 |
--|   1 |  SORT AGGREGATE    |         |     1 |     2 |            |          |
--|*  2 |   TABLE ACCESS FULL| ACTORES |  1172 |  2344 |    68   (0)| 00:00:01 |
--------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   2 - filter("SEXO"='H')

--4.3
--creamos el indice y las estadisticas
create index idx_sexo on ACTORES(SEXO);
exec dbms_stats.gather_table_stats(ownname => 'GIISGBD213', tabname => 'ACTORES', cascade => true);

--4.4
--volvemos a ejecutar los planes
explain plan for
select * from ACTORES where SEXO = 'H';
select * from table(dbms_xplan.display);

--Plan hash value: 2070918200
 
--------------------------------------------------------------------------------------------------
--| Id  | Operation                           | Name     | Rows  | Bytes | Cost (%CPU)| Time     |
--------------------------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT                    |          |  1382 |   145K|    28   (0)| 00:00:01 |
--|   1 |  TABLE ACCESS BY INDEX ROWID BATCHED| ACTORES  |  1382 |   145K|    28   (0)| 00:00:01 |
--|*  2 |   INDEX RANGE SCAN                  | IDX_SEXO |  1382 |       |     3   (0)| 00:00:01 |
--------------------------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   2 - access("SEXO"='H')
 
--Note
-------
 --  - SQL plan baseline "SQL_PLAN_5sagc315vquw02a6b4262" used for this statement

explain plan for
select count(*) from ACTORES where SEXO = 'H';
select * from table(dbms_xplan.display);

--Plan hash value: 3805308972
 
--------------------------------------------------------------------------------
--| Id  | Operation         | Name     | Rows  | Bytes | Cost (%CPU)| Time     |
--------------------------------------------------------------------------------
--|   0 | SELECT STATEMENT  |          |     1 |     2 |     3   (0)| 00:00:01 |
--|   1 |  SORT AGGREGATE   |          |     1 |     2 |            |          |
--|*  2 |   INDEX RANGE SCAN| IDX_SEXO |  1382 |  2764 |     3   (0)| 00:00:01 |
--------------------------------------------------------------------------------
 
--Predicate Information (identified by operation id):
-----------------------------------------------------
 
--   2 - access("SEXO"='H')


--4.5
--eliminamos el indice y creamos el indice bitmap
drop index idx_sexo;
create BITMAP index idx_sexo on ACTORES(SEXO);
exec dbms_stats.gather_table_stats(ownname => 'GIISGBD213', tabname => 'ACTORES', cascade => true);



