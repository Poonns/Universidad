package es.uv.bd.model;

import java.sql.*;
import java.text.ParseException;
import java.text.SimpleDateFormat;

public class RelacionDAO {
    public static final String DRIVER = "oracle.jdbc.OracleDriver";
    public static final String DBURL = "jdbc:oracle:thin:@pokemon.uv.es:1521:ORCL";
    public static final String USERNAME = "AL053"; 
    public static final String PASSWORD = "AL053";

    private static final String CREATE = 
            "INSERT INTO AL053_RELACION (id, desde, hasta, subtipo, al053_persona_id) " +
            "VALUES (?, TO_DATE(?,'DD-MM-YYYY'), TO_DATE(?,'DD-MM-YYYY'), ?, ?)";
    
    private static final String READ = 
            "SELECT id, TO_CHAR(desde,'DD-MM-YYYY') AS f_desde, " +
            "       TO_CHAR(hasta,'DD-MM-YYYY') AS f_hasta, " +
            "       subtipo, al053_persona_id " +
            "  FROM AL053_RELACION " +
            " WHERE id = ?";
    
    private static final String UPDATE =
            "UPDATE AL053_RELACION " +
            "   SET desde = TO_DATE(?,'DD-MM-YYYY'), " +
            "       hasta = TO_DATE(?,'DD-MM-YYYY'), " +
            "       subtipo = ?, al053_persona_id = ? " +
            " WHERE id = ?";
    
    private static final String DELETE =
            "DELETE FROM AL053_RELACION WHERE id = ?";

    public RelacionDAO() {}
   
    public void crearRelacion(Relacion relacion) throws 
            ClassNotFoundException, InstantiationException, IllegalAccessException, SQLException {
        
        SimpleDateFormat sdf = new SimpleDateFormat("dd-MM-yyyy");

        try (Connection oracleConn = DriverManager.getConnection(DBURL,USERNAME,PASSWORD)) {
            oracleConn.setAutoCommit(false);

            PreparedStatement insert = oracleConn.prepareStatement(CREATE);
            insert.setInt(1, relacion.getId());
            insert.setString(2, relacion.getDesde() != null ? sdf.format(relacion.getDesde()) : null);
            insert.setString(3, relacion.getHasta() != null ? sdf.format(relacion.getHasta()) : null);
            insert.setInt(4, relacion.getSubtipo());
            insert.setInt(5, relacion.getPersonaId());
            insert.executeUpdate();
            
            oracleConn.commit();
            oracleConn.setAutoCommit(true);
            oracleConn.close();
        }
    }

    
    public Relacion leerRelacion(int id) throws ClassNotFoundException, InstantiationException, IllegalAccessException, SQLException, ParseException {
        
        Relacion relacion = new Relacion();
        SimpleDateFormat sdf = new SimpleDateFormat("dd-MM-yyyy");
      
        try (Connection oracleConn = DriverManager.getConnection(DBURL,USERNAME,PASSWORD)) {
            PreparedStatement read = oracleConn.prepareStatement(READ);
            read.setInt(1, id);
            ResultSet rs = read.executeQuery();

            if (rs.next()) {
                relacion.setId(rs.getInt("id"));
                
                String fDesde = rs.getString("f_desde");
                if (fDesde != null) relacion.setDesde(sdf.parse(fDesde));
                
                String fHasta = rs.getString("f_hasta");
                if (fHasta != null) relacion.setHasta(sdf.parse(fHasta));
                
                relacion.setSubtipo(rs.getInt("subtipo"));
                relacion.setPersonaId(rs.getInt("al053_persona_id"));
            }
            return relacion;
        }
    }
    
    public void actualizarRelacion(Relacion relacion) throws ClassNotFoundException, 
           InstantiationException, IllegalAccessException, SQLException {
           
        SimpleDateFormat sdf = new SimpleDateFormat("dd-MM-yyyy");

        try (Connection oracleConn = DriverManager.getConnection(DBURL,USERNAME,PASSWORD)) {
            oracleConn.setAutoCommit(false);

            PreparedStatement update = oracleConn.prepareStatement(UPDATE);
            
            update.setString(1, relacion.getDesde() != null ? sdf.format(relacion.getDesde()) : null);
            update.setString(2, relacion.getHasta() != null ? sdf.format(relacion.getHasta()) : null);
            update.setInt(3, relacion.getSubtipo());
            update.setInt(4, relacion.getPersonaId());
            update.setInt(5, relacion.getId());
            update.executeUpdate();
            
            oracleConn.commit();
            oracleConn.setAutoCommit(true);
            oracleConn.close();
        }
    }
    
    public void borrarRelacion(int id) throws ClassNotFoundException, 
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