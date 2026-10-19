package es.uv.bd.model;

import java.sql.*;
import java.text.ParseException;
import java.text.SimpleDateFormat;
import java.util.Date;

public class PersonaDAO {
    public static final String DRIVER = "oracle.jdbc.OracleDriver";
    public static final String DBURL = "jdbc:oracle:thin:@pokemon.uv.es:1521:ORCL";
    public static final String USERNAME = "AL053"; 
    public static final String PASSWORD = "AL053";

    private static final String CREATE = 
            "INSERT INTO AL053_PERSONA (id, nombre, apellidos, nacimiento, fallecimiento) " + 
            "VALUES (?, ?, ?, TO_DATE(?,'DD-MM-YYYY'), TO_DATE(?,'DD-MM-YYYY'))";
    
    private static final String READ = 
            "SELECT id, nombre, apellidos, " +
            "       TO_CHAR(nacimiento,'DD-MM-YYYY') AS f_nacimiento, " +
            "       TO_CHAR(fallecimiento,'DD-MM-YYYY') AS f_fallecimiento " +
            "  FROM AL053_PERSONA " +
            " WHERE id = ?";
    
    private static final String UPDATE =
            "UPDATE AL053_PERSONA " +
            "   SET nombre = ?, apellidos = ?, " +
            "       nacimiento = TO_DATE(?,'DD-MM-YYYY'), " +
            "       fallecimiento = TO_DATE(?,'DD-MM-YYYY') " +
            " WHERE id = ?";
    
    private static final String DELETE =
            "DELETE FROM AL053_PERSONA WHERE id = ?";

    public PersonaDAO() {}
    
    public void crearPersona(Persona persona) throws 
            ClassNotFoundException, InstantiationException, IllegalAccessException, SQLException {
        
        SimpleDateFormat sdf = new SimpleDateFormat("dd-MM-yyyy");

        try (Connection oracleConn = DriverManager.getConnection(DBURL,USERNAME,PASSWORD)) {
            oracleConn.setAutoCommit(false); //utilizamos setAutoCommit(false) y commit para que si hay algo que falle no se guarde basura en la bbdd

            PreparedStatement insert = oracleConn.prepareStatement(CREATE);
            insert.setInt(1, persona.getId());
            insert.setString(2, persona.getNombre());
            insert.setString(3, persona.getApellidos());
            insert.setString(4, persona.getNacimiento() != null ? sdf.format(persona.getNacimiento()) : null);
            insert.setString(5, persona.getFallecimiento() != null ? sdf.format(persona.getFallecimiento()) : null);
            insert.executeUpdate();
            
            oracleConn.commit();
            oracleConn.setAutoCommit(true);
        }
    }
    
    public Persona leerPersona(int id) throws ClassNotFoundException, InstantiationException, IllegalAccessException, SQLException, ParseException {
        
        Persona persona = new Persona();
        SimpleDateFormat sdf = new SimpleDateFormat("dd-MM-yyyy");
      
        try (Connection oracleConn = DriverManager.getConnection(DBURL,USERNAME,PASSWORD)) {
            PreparedStatement read = oracleConn.prepareStatement(READ);
            read.setInt(1, id);
            ResultSet rs = read.executeQuery();

            if (rs.next()) {
                persona.setId(rs.getInt("id"));
                persona.setNombre(rs.getString("nombre"));
                persona.setApellidos(rs.getString("apellidos"));
                
                String fNac = rs.getString("f_nacimiento");
                if (fNac != null) persona.setNacimiento(sdf.parse(fNac));
                
                String fFal = rs.getString("f_fallecimiento");
                if (fFal != null) persona.setFallecimiento(sdf.parse(fFal));
            }
            return persona;
        }
    }

    public void actualizarPersona(Persona persona) throws ClassNotFoundException, 
           InstantiationException, IllegalAccessException, SQLException {
           
        SimpleDateFormat sdf = new SimpleDateFormat("dd-MM-yyyy");

        try (Connection oracleConn = DriverManager.getConnection(DBURL,USERNAME,PASSWORD)) {
            oracleConn.setAutoCommit(false);

            PreparedStatement update = oracleConn.prepareStatement(UPDATE);
            
            update.setString(1, persona.getNombre());
            update.setString(2, persona.getApellidos());
            update.setString(3, persona.getNacimiento() != null ? sdf.format(persona.getNacimiento()) : null);
            update.setString(4, persona.getFallecimiento() != null ? sdf.format(persona.getFallecimiento()) : null);
            update.setInt(5, persona.getId());
            update.executeUpdate();
            
            oracleConn.commit();
            oracleConn.setAutoCommit(true);
            oracleConn.close();
        }
    }

    public void borrarPersona(int id) throws ClassNotFoundException, 
           InstantiationException, IllegalAccessException, SQLException {

        try (Connection oracleConn = DriverManager.getConnection(DBURL,USERNAME,PASSWORD)) {
            oracleConn.setAutoCommit(false);
            
            PreparedStatement delete = oracleConn.prepareStatement(DELETE);
            delete.setInt(1, id);
            delete.executeUpdate();
            
            oracleConn.commit();
            oracleConn.setAutoCommit(true);
            oracleConn.close();
        }
    }
}