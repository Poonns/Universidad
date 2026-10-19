package es.uv.bd.test;

import es.uv.bd.model.Relacion;
import es.uv.bd.model.RelacionDAO;
import es.uv.bd.model.Persona;
import es.uv.bd.model.PersonaDAO;
import java.sql.SQLException;
import java.text.ParseException;
import java.text.SimpleDateFormat;

public class RelacionDAOTest {
    private static final String MODULO = "RelacionDaoTest::";
    private final RelacionDAO relacionDao;
    private final PersonaDAO personaDao;
    private Relacion relPrueba, leida;
    private Persona personaAuxiliar;
    
    public static void main(String[] args) {
        RelacionDAOTest test = new RelacionDAOTest();
        test.doTest();
    }

    public RelacionDAOTest() {
        relacionDao = new RelacionDAO();
        personaDao = new PersonaDAO(); 
    }
    
    public void doTest() {
        SimpleDateFormat dateFormat = new SimpleDateFormat("dd-MM-yyyy");
            
        try {
            personaAuxiliar = new Persona(800, "Aux", "Para Relacion", dateFormat.parse("01-01-2000"), null);
            personaDao.crearPersona(personaAuxiliar);

            print("Test 1: Crear relacion");            
            relPrueba = new Relacion(100, dateFormat.parse("01-01-2022"), dateFormat.parse("31-12-2023"), 1, personaAuxiliar.getId());
            relacionDao.crearRelacion(relPrueba);
            print("Test 1: " + relPrueba + " creada.");
            
            print("Test 2: Buscar relacion");
            leida = relacionDao.leerRelacion(relPrueba.getId());
            print("Test 2: " + leida + " leída.");
            
            print("Test 3: Cambiar subtipo de la relacion");
            leida.setSubtipo(2);
            relacionDao.actualizarRelacion(leida);
            relPrueba = relacionDao.leerRelacion(leida.getId());
            print("Test 3: " + relPrueba + " modificada.");
            
            print("Test 4: Borrar relacion " + relPrueba.getId());
            relacionDao.borrarRelacion(relPrueba.getId());
            print("Test 4: Relacion borrada");

            personaDao.borrarPersona(personaAuxiliar.getId());
            
            print("Todos los test correctos");
            
        } catch (ParseException | ClassNotFoundException | InstantiationException | IllegalAccessException | SQLException ex) {
            print("Error ejecutando tests");
            print(" Excepción: " + ex.getMessage());
            print("");
            ex.printStackTrace();
        }
    }
    
    private void print(String string) {
        System.out.println(MODULO + string);
    }
}