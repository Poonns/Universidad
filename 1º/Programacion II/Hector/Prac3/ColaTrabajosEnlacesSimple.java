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

        TrabajoImpresora borrado = prim;
        prim = prim.getSiguiente();
        talla --;

        if(prim == null){
            ult = null;
        }

        return borrado;
    }

    public TrabajoImpresora consultarTrabajo() throws NoSuchElementException{
        if(prim == null){
            throw new NoSuchElementException("La lista está vacía");
        }
         return prim;
    }

    public int getTalla(){
        return talla;
    }

    public boolean esVacia(){
        if(prim == null){
            return false;
        }
        return true;
    }

    public String toString(){
        String cadena = "[";
        for(TrabajoImpresora actual = prim; actual != null; actual = actual.getSiguiente()){
            cadena += actual.toString() + ",";
        }
        cadena += "]";

        return cadena;
    }

}
