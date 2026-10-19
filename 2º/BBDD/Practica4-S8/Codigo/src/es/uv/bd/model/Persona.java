
package es.uv.bd.model;

import java.text.SimpleDateFormat;
import java.util.Date;

public class Persona {

    private int id;
    private String nombre;
    private String apellidos;
    private Date nacimiento;
    private Date fallecimiento;

    public Persona() {}
    //constructor
    public Persona(int id, String nombre, String apellidos, Date nacimiento, Date fallecimiento) {
        this.id = id;
        this.nombre = nombre;
        this.apellidos = apellidos;
        this.nacimiento = nacimiento;
        this.fallecimiento = fallecimiento;
    }
    
    public int getId() { return id; }
    public void setId(int id) { this.id = id; }

    public String getNombre() { return nombre; }
    public void setNombre(String nombre) { this.nombre = nombre; }

    public String getApellidos() { return apellidos; }
    public void setApellidos(String apellidos) { this.apellidos = apellidos; }

    public Date getNacimiento() { return nacimiento; }
    public void setNacimiento(Date nacimiento) { this.nacimiento = nacimiento; }

    public Date getFallecimiento() { return fallecimiento; }
    public void setFallecimiento(Date fallecimiento) { this.fallecimiento = fallecimiento; }

    @Override
    public String toString() {
        SimpleDateFormat sdf = new SimpleDateFormat("dd-MM-yyyy");
        String fNac = (nacimiento != null) ? sdf.format(nacimiento) : "null";
        String fFal = (fallecimiento != null) ? sdf.format(fallecimiento) : "null";
        
        return "Persona{" + "id=" + id + ", nombre=" + nombre + ", apellidos=" + apellidos + ", nacimiento=" + fNac + ", fallecimiento=" + fFal + '}';
    }
}