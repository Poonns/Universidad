CREATE TABLE SUGERENCIAS (
    codcoche VARCHAR2(10),
    cifc_origen VARCHAR2(10),
    cifc_destino VARCHAR2(10),
    CONSTRAINT PK_SUGERENCIAS PRIMARY KEY (codcoche, cifc_origen)
);

CREATE OR REPLACE PROCEDURE CREA_SUGERENCIAS IS
    CURSOR c_candidatos IS
        SELECT d.cifc, d.codcoche
        FROM DISTRIBUCION d
        WHERE NOT EXISTS (
            SELECT 1 
            FROM VENTAS v 
            WHERE v.cifc = d.cifc AND v.codcoche = d.codcoche
        ); 

    v_cifc_destino SUGERENCIAS.cifc_destino%TYPE; 

BEGIN
    DELETE FROM SUGERENCIAS;

    FOR r_car IN c_candidatos LOOP 
        BEGIN
            SELECT cifc INTO v_cifc_destino
            FROM (
                SELECT cifc, COUNT(*) as total_ventas
                FROM VENTAS
                WHERE codcoche = r_car.codcoche
                GROUP BY cifc
                ORDER BY total_ventas DESC 
            )
            WHERE ROWNUM = 1;

            INSERT INTO SUGERENCIAS (codcoche, cifc_origen, cifc_destino)
            VALUES (r_car.codcoche, r_car.cifc, v_cifc_destino);

        EXCEPTION
            WHEN NO_DATA_FOUND THEN
                INSERT INTO SUGERENCIAS (codcoche, cifc_origen, cifc_destino)
                VALUES (r_car.codcoche, r_car.cifc, NULL);
            WHEN OTHERS THEN
                NULL; 
        END;
    END LOOP;

    COMMIT;

EXCEPTION
    WHEN OTHERS THEN
        ROLLBACK;
END CREA_SUGERENCIAS;