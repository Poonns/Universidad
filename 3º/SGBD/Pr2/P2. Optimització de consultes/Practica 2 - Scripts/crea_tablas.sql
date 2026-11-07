drop table actuacion;

drop table peliculas;

drop table actores;

CREATE TABLE actores (
  oid decimal(10,0) NOT NULL,
  nom char(100) NOT NULL,
  sexo char(1) NOT NULL,
  PRIMARY KEY (oid)
);

CREATE TABLE peliculas (
  oid decimal(10,0) NOT NULL,
  titulo char(200) NOT NULL,
  anyo decimal(4,0) NOT NULL,
  PRIMARY KEY (oid)
);

CREATE TABLE actuacion (
  actor decimal(10,0) NOT NULL,
  peli decimal(10,0) NOT NULL,
  papel char(100) NOT NULL,
  PRIMARY KEY (actor,peli,papel)
);
