package es.uv.bd.test;

import es.uv.bd.model.Persona;
import es.uv.bd.model.PersonaDAO;
import java.sql.SQLException;
import java.text.ParseException;
import java.text.SimpleDateFormat;

public class PersonaDAOTest {
    private static final String MODULO = "PersonaDaoTest::";
    private final PersonaDAO personaDao;
    private Persona personaPrueba, leida;
    
    public static void main(String[] args) {
        PersonaDAOTest test = new PersonaDAOTest();
        test.doTest();
    }

    public PersonaDAOTest() {
        personaDao = new PersonaDAO();
    }
    
    public void doTest() {
        SimpleDateFormat dateFormat = new SimpleDateFormat("dd-MM-yyyy");
            
        try {
            print("Test 1: Crear persona");            
            personaPrueba = new Persona(900, "Juan", "Perez", dateFormat.parse("15-05-1990"), null);
            personaDao.crearPersona(personaPrueba);
            print("Test 1: " + personaPrueba + " creada.");
            
            print("Test 2: Buscar persona");
            leida = personaDao.leerPersona(personaPrueba.getId());
            print("Test 2: " + leida + " leída.");
            
            print("Test 3: Cambiar nombre de la persona");
            leida.setNombre("Juan Antonio");
            personaDao.actualizarPersona(leida);
            personaPrueba = personaDao.leerPersona(leida.getId());
            print("Test 3: " + personaPrueba + " modificada.");
            
            print("Test 4: Borrar persona " + personaPrueba.getId() + " (" + personaPrueba.getNombre() + ")");
            personaDao.borrarPersona(personaPrueba.getId());
            print("Test 4: Persona borrada");
            
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