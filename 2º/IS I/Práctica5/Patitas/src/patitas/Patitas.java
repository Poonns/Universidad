/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Main.java to edit this template
 */
package patitas;
import java.util.Date;
import modelo.*;
import java.util.ArrayList;
import vista.*;
import javax.swing.UIManager;

/**
 *
 * @author Carla Terol <Carla Terol>
 */
public class Patitas {

    /**
     * @param args the command line arguments
     */
    public static void main(String[] args) {
       
        // TODO code application logic here
        
        
        
        
        
        CentroAdopcion patitas = new CentroAdopcion();
        
        //Creem una llista de animals per a fer proves
        ArrayList<Animal> listaAnimales = new ArrayList<Animal>();
        
        listaAnimales.add(new Animal("1", "Perro", "Chihuahua", 2, new Date(), "Muy nervioso pero cariñoso", patitas, new DisponibleAdopcion()));
        listaAnimales.add(new Animal("2", "Gato", "Siamés", 4, new Date(), "Le gusta dormir en el teclado", patitas, new DisponibleAdopcion()));
        listaAnimales.add(new Animal("3", "Perro", "Pastor Alemán", 5, new Date(), "Muy guardián y obediente", patitas, new ProcesoAdopcion()));
        listaAnimales.add(new Animal("4", "Gato", "Persa", 1, new Date(), "Pelo largo y muy tranquilo", patitas, new ProcesoAdopcion()));
        listaAnimales.add(new Animal("5", "Conejo", "Enano", 1, new Date(), "Le encantan las zanahorias", patitas, new DisponibleAdopcion()));
        listaAnimales.add(new Animal("6", "Perro", "Golden Retriever", 3, new Date(), "El mejor amigo de los niños", patitas, new DisponibleAdopcion()));
        listaAnimales.add(new Animal("7", "Gato", "Maine Coon", 6, new Date(), "Es gigante y muy noble", patitas, new DisponibleAdopcion()));
        listaAnimales.add(new Animal("8", "Perro", "Bulldog Francés", 2, new Date(), "Ronca un poco pero es un amor", patitas, new DisponibleAdopcion()));
        listaAnimales.add(new Animal("9", "Pájaro", "Agaporni", 1, new Date(), "No para de piar en todo el día", patitas, new NoAdoptable()));
        listaAnimales.add(new Animal("10", "Perro", "Beagle", 4, new Date(), "Gran olfato, siempre buscando comida", patitas, new DisponibleAdopcion()));
        listaAnimales.add(new Animal("11", "Gato", "Común Europeo", 3, new Date(), "Cazador experto de moscas", patitas, new NoAdoptable()));
        listaAnimales.add(new Animal("12", "Perro", "Labrador", 7, new Date(), "Le encanta el agua y jugar", patitas, new DisponibleAdopcion()));
        listaAnimales.add(new Animal("13", "Hámster", "Ruso", 1, new Date(), "Corre en la rueda por las noches", patitas, new NoAdoptable()));
        listaAnimales.add(new Animal("14", "Perro", "Border Collie", 2, new Date(), "Demasiado inteligente, te vigila", patitas, new DisponibleAdopcion()));
        listaAnimales.add(new Animal("15", "Gato", "Bengala", 3, new Date(), "Parece un leopardo en miniatura", patitas, new ProcesoAdopcion()));
        
        patitas.setAnimales(listaAnimales);
        
        //Creamos un adoptante y lo metemos en lista de adoptantes del centro
        
        Adoptante adoptante = new Adoptante("Pau", "20101010K", "Plaza España", "Piso", "Experiencia con perros", patitas);
        
        
        ArrayList<Adoptante> listaAdoptantes = new ArrayList<Adoptante>();
        listaAdoptantes.add(adoptante);
        
        patitas.setAdoptantes(listaAdoptantes);
        
        javax.swing.SwingUtilities.invokeLater(() -> {
            VentanaConsultarAnimales frame = new VentanaConsultarAnimales(adoptante);
            frame.setVisible(true);
        });
        
        
        
        
    }
    
}
