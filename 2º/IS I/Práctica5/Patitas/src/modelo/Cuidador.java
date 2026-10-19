package modelo;
import java.util.Vector;

public class Cuidador extends Personal {
	public Responsable _Responsable;
	public Vector<Confirmada> _confirmada = new Vector<Confirmada>();

	public Cuidador() {
		throw new UnsupportedOperationException();
	}
}