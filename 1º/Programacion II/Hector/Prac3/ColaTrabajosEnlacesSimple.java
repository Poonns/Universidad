package Hector.Prac3;

import java.util.NoSuchElementException;

public class ColaTrabajosEnlacesSimple implements ColaTrabajos {

    private int talla;
    private TrabajoImpresora  prim;
    private TrabajoImpresora ult;

    public ColaTrabajosEnlacesSimple(){
        this.talla = 0;
        this.prim = null;
        this.ult = null;
    }

    public void insertar(TrabajoImpresora t){
        
        if(talla == 0){
            prim = t;
            ult = t;
        } else {
            ult.setSiguiente(t);
            ult = t;
        }
        talla ++;
    }

    public TrabajoImpresora extraerPrimero() throws NoSuchElementException{
        if(prim == null){
            throw new NoSuchElementException("No hay elementos");
        }

        prim = prim.getSiguiente();
    }

}
