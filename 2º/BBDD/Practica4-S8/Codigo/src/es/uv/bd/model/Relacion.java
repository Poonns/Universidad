package es.uv.bd.model;

import java.text.SimpleDateFormat;
import java.util.Date;

public class Relacion {

    private int id;
    private Date desde;
    private Date hasta;
    private int subtipo;
    private int personaId;

    public Relacion() {}
    //constructor
    public Relacion(int id, Date desde, Date hasta, int subtipo, int personaId) {
        this.id = id;
        this.desde = desde;
        this.hasta = hasta;
        this.subtipo = subtipo;
        this.personaId = personaId;
    }
    
    public int getId() { 
        return id;
    }
    public void setId(int id) { 
        this.id = id;
    }

    public Date getDesde() { 
        return desde;
    }
    public void setDesde(Date desde) { 
        this.desde = desde;
    }

    public Date getHasta() { 
        return hasta;
    }
    public void setHasta(Date hasta) { 
        this.hasta = hasta;
    }

    public int getSubtipo() { 
        return subtipo;
    }
    public void setSubtipo(int subtipo) { 
        this.subtipo = subtipo;
    }

    public int getPersonaId() { 
        return personaId;
    }
    public void setPersonaId(int personaId) { 
        this.personaId = personaId;
    }
    
    @Override
    public String toString() {
        SimpleDateFormat sdf = new SimpleDateFormat("dd-MM-yyyy");
        String fDesde = (desde != null) ? sdf.format(desde) : "null";
        String fHasta = (hasta != null) ? sdf.format(hasta) : "null";
        
        return "Relacion{" + "id=" + id + ", desde=" + fDesde + ", hasta=" + fHasta + ", subtipo=" + subtipo + ", personaId=" + personaId + '}';
    }
}

